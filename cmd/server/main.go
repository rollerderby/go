package main

//go:generate $GOPATH/bin/esc -o html.go ../../html

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"

	"github.com/rollerderby/go/auth"
	"github.com/rollerderby/go/entity"
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/ruleset"
	"github.com/rollerderby/go/state"
	"github.com/rollerderby/go/websocket"
)

var log = logger.New("main")

// openBrowser tries to open the URL in a browser,
// and returns whether it succeed in doing so.
func openBrowser(url string) bool {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

func setVersion() {
	state.Root.Lock()
	defer state.Root.Unlock()

	versionStr := state.NewString().(*state.String)
	versionStr.SetValue(version)

	state.Root.Add("Version", "", versionStr)
}

func httpLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("%s: %s %s %s", r.Host, r.RemoteAddr, r.Method, r.URL)
		w.Header().Set("cache-control", "private, max-age=0, no-cache")
		handler.ServeHTTP(w, r)
	})
}

func extractHTMLFolder() error {
	for key, value := range _escData {
		if value.isDir && key != "/" {
			local := path.Join(".", key)
			if _, err := os.Stat(local); os.IsNotExist(err) {
				log.Debugf("Create missing HTML directory: %v", key)
				if err := os.MkdirAll(local, 0775); err != nil {
					return err
				}
			}
		}
	}
	for key, value := range _escData {
		if !value.isDir {
			local := path.Join(".", key)
			if _, err := os.Stat(local); os.IsNotExist(err) {
				log.Debugf("Create missing HTML file: %v", key)
				if data, err := FSByte(false, key); err != nil {
					return err
				} else if err := ioutil.WriteFile(local, data, 0664); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func initializeWebserver(port uint16, signals chan os.Signal) error {
	if err := extractHTMLFolder(); err != nil {
		log.Errorf("Cannot extract HTML folder: %v", err)
		return err
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", http.FileServer(http.Dir("html")))
	go func() {
		log.Infof("Starting up http server on :%v", port)
		log.Crit(http.ListenAndServe(fmt.Sprintf(":%d", port), httpLog(httpMux)))
		signals <- os.Kill
	}()

	return nil
}

func main() {
	log.Info("Initializing")

	port := flag.Int("port", 8000, "Port to listen on")
	verbose := flag.Bool("v", false, "Print debugging information")
	flag.Parse()

	if *verbose {
		logger.SetLevel(logger.DEBUG)
	}

	setVersion()

	ruleset.Initialize()
	entity.Initialize()
	auth.Initialize()
	websocket.Initialize()

	state.Root.LoadSavedConfigs()

	if err := auth.AddGlobalUsers(); err != nil {
		log.Errorf("Unable to add global admin user: %v", err)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	if err := initializeWebserver(uint16(*port), signals); err != nil {
		return
	}
	go state.Root.SaveLoop()

	openBrowser(fmt.Sprintf("http://localhost:%v", uint16(*port)))

	signal.Notify(signals, os.Interrupt, os.Kill)
	s := <-signals
	log.Noticef("Server received signal: %v.  Shutting down", s)
}
