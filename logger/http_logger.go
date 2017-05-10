package logger

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

type transportLogger struct {
	rt   http.RoundTripper
	name string
	log  *stdlog.Logger
	c    int64
}

var initOnce sync.Once

func httpLoggerInit() {
	os.RemoveAll("logs")
	os.MkdirAll("logs", 0755)
}

func NewTransportLogger(name string, rt http.RoundTripper) *transportLogger {
	initOnce.Do(httpLoggerInit)

	if rt == nil {
		rt = http.DefaultTransport
	}
	tl := &transportLogger{rt: rt, name: name}

	f, err := os.Create(fmt.Sprintf("logs/%s.log", name))
	if err != nil {
		undef.log.Errorf("Cannot create http logs: %v", err)
	} else {
		tl.log = stdlog.New(f, "", stdlog.LstdFlags)
	}

	return tl
}

func (tl *transportLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	var b bytes.Buffer
	var err error
	logFile := &b

	tl.c++
	logFileName := fmt.Sprintf("logs/%s_%05d.log", tl.name, tl.c)

	defer func() {
		if tl.log == nil {
			undef.log.Debug(b.String())
			return
		}
		f, err := os.Create(logFileName)
		if err != nil {
			undef.log.Error(err)
			return
		}
		fmt.Fprint(f, b.String())
		f.Close()
	}()

	req.Body, err = httpLog(logFile, fmt.Sprintf("%v %v", req.Method, req.URL), req.Header, req.Body)
	if err != nil {
		return nil, err
	}
	res, err := tl.rt.RoundTrip(req)

	if err != nil {
		tl.log.Printf("%v  %v %v %v", logFileName, req.Method, req.URL, err)
	} else {
		tl.log.Printf("%v  %v %v %v", logFileName, req.Method, req.URL, res.Status)
		res.Body, err = httpLog(logFile, "RESPONSE", res.Header, res.Body)
		if err != nil {
			return nil, err
		}
	}

	return res, err
}

func httpLog(w io.Writer, headerText string, header http.Header, r io.ReadCloser) (io.ReadCloser, error) {
	fmt.Fprintf(w, "%v\n", headerText)

	if len(header) == 0 {
		fmt.Fprint(w, "\tNo Headers\n")
	} else {
		fmt.Fprint(w, "\tHeaders\n")
		var keys []string
		for key, _ := range header {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Fprintf(w, "\t\t%v: %q\n", key, header[key])
		}
	}

	if r == nil {
		fmt.Fprint(w, "\tNo Body\n\n")
		return nil, nil
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	r.Close()
	dataStr := string(data)
	fmt.Fprint(w, "\tBody\n")
	fmt.Fprintf(w, "\t\t%v\n\n", strings.Replace(dataStr, "\n", "\n\t\t", -1))

	return &BufferCloser{bytes.NewBuffer(data)}, nil
}

type BufferCloser struct {
	buf *bytes.Buffer
}

func (bc *BufferCloser) Read(p []byte) (n int, err error) {
	return bc.buf.Read(p)
}

func (bc *BufferCloser) Close() error {
	bc.buf = nil
	return nil
}
