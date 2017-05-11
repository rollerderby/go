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

var (
	log     = logger.New("server")
	restart *RestartSignal
)

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

type RestartSignal struct {
	ResetConfig bool
	ResetHTML   bool
}

func (rs *RestartSignal) String() string {
	return fmt.Sprintf("Restart { ResetConfig: %v  ResetHTML: %v }", rs.ResetConfig, rs.ResetHTML)
}

func (rs *RestartSignal) Signal() {
}

// Returns true if caller should restart
func Run(port uint16) bool {
	log.Info("Initializing")
	restart = &RestartSignal{}

	setVersion()

	signals := make(chan os.Signal, 1)
	if err := initializeWebserver(port, signals); err != nil {
		log.Errorf("Unable to initialize webserver: %v", err)
		return false
	}

	ruleset.Initialize()
	entity.Initialize()
	auth.Initialize()
	websocket.Initialize()

	state.Root.LoadSavedConfigs()

	if err := auth.AddGlobalUsers(); err != nil {
		log.Errorf("Unable to add global admin user: %v", err)
		return false
	}

	go state.Root.SaveLoop()

	openBrowser(fmt.Sprintf("http://localhost:%v", port))

	signal.Notify(signals, os.Interrupt, os.Kill, restart)
	s := <-signals

	if s, ok := s.(*RestartSignal); ok {
		log.Alertf("Restarting server: %+v", s)
		if s.ResetConfig {
			os.RemoveAll("config")
		}
		if s.ResetHTML {
			os.RemoveAll("html")
		}

		return true
	}

	log.Alertf("Server received signal: %v.  Shutting down", s)
	return false
}
