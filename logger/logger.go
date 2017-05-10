package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Level uint8

const (
	EMERG   Level = iota // system is unusable
	ALERT                // action must be taken immediately
	CRIT                 // critical conditions
	ERR                  // error conditions
	WARNING              // warning conditions
	NOTICE               // normal, but significant, condition
	INFO                 // informational message
	DEBUG                // debug-level message
)

var paddedLevels = map[Level]string{
	EMERG:   " EMERG   ",
	ALERT:   " ALERT   ",
	CRIT:    " CRIT    ",
	ERR:     " ERR     ",
	WARNING: " WARNING ",
	NOTICE:  " NOTICE  ",
	INFO:    " INFO    ",
	DEBUG:   " DEBUG   ",
}

func (l Level) String() string {
	switch l {
	case EMERG:
		return "EMERG"
	case ALERT:
		return "ALERT"
	case CRIT:
		return "CRIT"
	case ERR:
		return "ERR"
	case WARNING:
		return "WARNING"
	case NOTICE:
		return "NOTICE"
	case INFO:
		return "INFO"
	}

	// DEBUG and everything else
	return "DEBUG"
}

type Logger struct {
	name   string
	level  Level
	parent *Logger
}

type masterLogger struct {
	sync.Mutex
	logger  *log.Logger
	once    sync.Once
	out     io.WriteCloser
	outPath string
	outTee  bool
}

type undefLogger struct {
	buf []byte
	log *Logger
}

const UNKNOWN = "UNKNOWN"

var ml *masterLogger = &masterLogger{}
var undef *undefLogger = &undefLogger{log: &Logger{name: UNKNOWN, level: DEBUG}}
var RootLoggers []*Logger

func (u *undefLogger) logLine() bool {
	for pos, r := range string(u.buf) {
		if r == '\n' {
			u.log.Info(string(u.buf[:pos]))
			if pos == len(u.buf)-1 {
				u.buf = nil
			} else {
				u.buf = u.buf[pos+1:]
			}
			return true
		}
	}
	return false
}

func (u *undefLogger) Write(p []byte) (int, error) {
	u.buf = append(u.buf, p...)
	for {
		if !u.logLine() {
			break
		}
	}

	return len(p), nil
}

func init() {
	ml.once.Do(ml.init)
}

func (ml *masterLogger) init() {
	log.SetOutput(undef)
	log.SetFlags(0)

	ml.logger = log.New(os.Stdout, "", log.LstdFlags)
	RootLoggers = append(RootLoggers, undef.log)
}

func (ml *masterLogger) tryOpen(path string, tee bool) error {
	ml.once.Do(ml.init)

	var f io.WriteCloser
	var err error

	if f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return err
	}

	ml.Lock()
	defer ml.Unlock()

	if ml.out != nil {
		ml.out.Close()
	}
	ml.out = f
	ml.outPath = path
	ml.outTee = tee

	if tee {
		ml.logger.SetOutput(io.MultiWriter(f, os.Stdout))
	} else {
		ml.logger.SetOutput(f)
	}

	return nil
}

func (ml *masterLogger) closeAndOpen() error {
	ml.once.Do(ml.init)

	if ml.out != nil {
		return ml.tryOpen(ml.outPath, ml.outTee)
	}
	return nil
}

func (ml *masterLogger) close() {
	ml.once.Do(ml.init)

	ml.Lock()
	defer ml.Unlock()

	if ml.out != nil {
		ml.out.Close()
		ml.out = nil
		ml.outPath = ""
		ml.outTee = false
		ml.logger.SetOutput(os.Stdout)
	}
}

func (ml *masterLogger) log(msg string) {
	ml.once.Do(ml.init)

	ml.Lock()
	defer ml.Unlock()

	ml.logger.Print(msg)
}

func SetOutputFile(path string, tee bool) error {
	return ml.tryOpen(path, tee)
}

func CloseAndOpen() error {
	return ml.closeAndOpen()
}

func Close() {
	ml.close()
}

func New(name string) *Logger {
	l := &Logger{name: name, level: INFO}
	RootLoggers = append(RootLoggers, l)
	return l
}

var bufPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return new(bytes.Buffer)
	},
}

func (l *Logger) getBuffer(level Level) *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()

	levelStr, ok := paddedLevels[level]
	if !ok {
		levelStr = paddedLevels[DEBUG]
	}
	b.WriteString(levelStr)
	b.WriteString("  ")
	b.WriteString(l.fullName())
	b.WriteString(": ")
	return b
}

func (l *Logger) fullName() string {
	name := l.name
	for {
		if l.parent == nil {
			break
		}
		l = l.parent
		name = l.name + "." + name
	}

	return name
}

func (l *Logger) Name() string {
	return l.fullName()
}

func (l *Logger) Level() Level {
	level := l.level
	for {
		if l.parent == nil {
			break
		}
		l = l.parent
		if l.level > level {
			level = l.level
		}
	}

	return level
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) Child(name string) *Logger {
	cl := &Logger{name: name, level: l.level, parent: l}
	return cl
}

func (l *Logger) Print(level Level, v ...interface{}) {
	if level > l.Level() {
		return
	}

	b := l.getBuffer(level)
	defer bufPool.Put(b)

	fmt.Fprint(b, v...)
	ml.log(b.String())

	if level == EMERG {
		os.Exit(1)
	}
}

func (l *Logger) Printf(level Level, format string, v ...interface{}) {
	if level > l.Level() {
		return
	}

	b := l.getBuffer(level)
	defer bufPool.Put(b)

	fmt.Fprintf(b, format, v...)
	ml.log(b.String())

	if level == EMERG {
		os.Exit(1)
	}
}

func (l *Logger) Println(level Level, v ...interface{}) {
	if level > l.Level() {
		return
	}

	b := l.getBuffer(level)
	defer bufPool.Put(b)

	fmt.Fprintln(b, v...)
	ml.log(b.String())

	if level == EMERG {
		os.Exit(1)
	}
}

func (l *Logger) Emerg(v ...interface{})   { l.Print(EMERG, v...) }
func (l *Logger) Fatal(v ...interface{})   { l.Print(EMERG, v...) }
func (l *Logger) Alert(v ...interface{})   { l.Print(ALERT, v...) }
func (l *Logger) Crit(v ...interface{})    { l.Print(CRIT, v...) }
func (l *Logger) Err(v ...interface{})     { l.Print(ERR, v...) }
func (l *Logger) Error(v ...interface{})   { l.Print(ERR, v...) }
func (l *Logger) Warning(v ...interface{}) { l.Print(WARNING, v...) }
func (l *Logger) Notice(v ...interface{})  { l.Print(NOTICE, v...) }
func (l *Logger) Info(v ...interface{})    { l.Print(INFO, v...) }
func (l *Logger) Debug(v ...interface{})   { l.Print(DEBUG, v...) }

func (l *Logger) Emergf(format string, v ...interface{})   { l.Printf(EMERG, format, v...) }
func (l *Logger) Fatalf(format string, v ...interface{})   { l.Printf(EMERG, format, v...) }
func (l *Logger) Alertf(format string, v ...interface{})   { l.Printf(ALERT, format, v...) }
func (l *Logger) Critf(format string, v ...interface{})    { l.Printf(CRIT, format, v...) }
func (l *Logger) Errf(format string, v ...interface{})     { l.Printf(ERR, format, v...) }
func (l *Logger) Errorf(format string, v ...interface{})   { l.Printf(ERR, format, v...) }
func (l *Logger) Warningf(format string, v ...interface{}) { l.Printf(WARNING, format, v...) }
func (l *Logger) Noticef(format string, v ...interface{})  { l.Printf(NOTICE, format, v...) }
func (l *Logger) Infof(format string, v ...interface{})    { l.Printf(INFO, format, v...) }
func (l *Logger) Debugf(format string, v ...interface{})   { l.Printf(DEBUG, format, v...) }
