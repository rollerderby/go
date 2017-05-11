package server

//go:generate $GOPATH/bin/esc -pkg server -o html.go ../html

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/rollerderby/go/auth"
	"github.com/rollerderby/go/entity"
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/ruleset"
	"github.com/rollerderby/go/state"
	"github.com/rollerderby/go/websocket"
)

var log = logger.New("server")

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

func Run(port uint16) {
	log.Info("Initializing")

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
	if err := initializeWebserver(port, signals); err != nil {
		return
	}
	go state.Root.SaveLoop()

	openBrowser(fmt.Sprintf("http://localhost:%v", port))

	signal.Notify(signals, os.Interrupt, os.Kill)
	s := <-signals
	log.Noticef("Server received signal: %v.  Shutting down", s)
}
