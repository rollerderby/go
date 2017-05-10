package main

//go:generate $GOPATH/bin/esc -prefix html -o html.go ../../html

import (
	"flag"
	"os/exec"
	"runtime"
	"time"

	"github.com/rollerderby/go/entity"
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/ruleset"
	"github.com/rollerderby/go/state"
	"github.com/rollerderby/go/websocket"
)

var log = logger.New("main")

func addDummyTeams(names ...string) {
	state.Root.Lock()
	defer state.Root.Unlock()

	for _, name := range names {
		if t, err := entity.Teams.New(""); err != nil {
			log.Errorf("Cannot create team: %v", err)
		} else {
			t.SetName(name)
		}
	}
}

func addDummyPerson() {
	state.Root.Lock()
	defer state.Root.Unlock()

	per, err := entity.People.New("")
	if err != nil {
		log.Errorf("Cannot create person: %v", err)
		return
	}
	per.SetName("Michael Mitton")
	cert, err := per.Certs().New("")
	if err != nil {
		log.Errorf("Cannot create cert: %v", err)
	}
	cert.SetOrganization("WFTDA")
	cert.SetType("NSO 2")
}

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

func main() {
	log.Info("Initializing")

	addTeams := flag.Bool("addTeams", false, "Add Dummy Teams")
	addPerson := flag.Bool("addPerson", false, "Add Dummy Person")
	flag.Parse()

	setVersion()

	ruleset.Initialize()
	entity.Initialize()
	websocket.Initialize()

	// addDummyTeams("Team A", "Team B")

	state.Root.LoadSavedConfigs()

	if *addTeams {
		teamsNeeded := map[string]bool{
			"Team A": true,
			"Team B": true,
			"Team C": true,
			"Team D": true,
			"Team E": true,
		}

		for _, t := range entity.Teams.Values() {
			delete(teamsNeeded, t.Name())
		}

		for key, _ := range teamsNeeded {
			addDummyTeams(key)
		}
	}

	if *addPerson {
		found := false
		for _, p := range entity.People.Values() {
			if p.Name() == "Michael Mitton" {
				found = true
				break
			}
		}
		if !found {
			addDummyPerson()
		}
	}

	go state.Root.SaveLoop()

	openBrowser("http://localhost:8000")
	time.Sleep(15 * time.Second)
}
