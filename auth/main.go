package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/rollerderby/go/entity"
	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/state"
)

//go:generate $GOPATH/bin/buildStates

var (
	log        = logger.New("Auth")
	rsaKey     *rsa.PrivateKey
	fileserver http.Handler
)

func GenerateKey() error {
	var err error
	rsaKey, err = rsa.GenerateKey(rand.Reader, 2048)

	return err
}

func Initialize(mux *ServeMux) error {
	log.Info("Initializing")

	initializeState()

	if err := GenerateKey(); err != nil {
		return err
	}

	fileserver = http.StripPrefix("/auth/", http.FileServer(http.Dir("html/auth")))
	mux.HandleFunc("", "/auth/", authHandler, nil)
	mux.HandleFunc("", "/auth/logout/", logoutHandler, nil)

	return nil
}

func AddGlobalUsers() error {
	state.Root.Lock()
	defer state.Root.Unlock()

	defaultUsers := []struct {
		username string
		password string
		isSuper  bool
		groups   []string
	}{
		{"admin", "admin", true, nil},
		{"readonly", "readonly", false, []string{"readonly"}},
	}

	var err error

	for _, u := range defaultUsers {
		var person *entity.PersonHelper

		for _, per := range entity.People.FindByName(u.username) {
			person = per
			break
		}

		if person == nil {
			person, err = entity.People.New("")
			if err != nil {
				log.Errorf("Cannot create person: %v", err)
				return err
			}
			person.SetName(u.username)
		}

		users := Users.FindByPersonID(person.ID())
		if len(users) == 0 {
			if _, err := Users.AddUser(u.username, u.password, u.isSuper, u.groups, person.ID()); err != nil {
				return err
			}
		}
	}
	return nil
}

func clearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "auth", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "auth_sig", MaxAge: -1, Path: "/"})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearCookies(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	var redir string
	if redir = r.FormValue("redir"); redir != "" {
		http.SetCookie(w, &http.Cookie{Name: "auth_redirect", Value: redir, MaxAge: 0, Path: "/"})
	} else if redir_cookie, _ := r.Cookie("auth_redirect"); redir_cookie != nil && strings.TrimSpace(redir_cookie.Value) != "" {
		redir = redir_cookie.Value
	} else {
		redir = r.Referer()
		http.SetCookie(w, &http.Cookie{Name: "auth_redirect", Value: redir, MaxAge: 0, Path: "/"})
	}
	if redir == "" || strings.HasPrefix(redir, "/auth/") {
		redir = "/"
	}

	if user := CheckAuth(r); user == nil {
		// Not authenticated, check for login attempt
		switch r.URL.Path {
		case "/auth/":
			username := r.PostFormValue("username")
			password := r.PostFormValue("password")
			if username != "" || password != "" {
				if user, err := Authenticate(w, r, username, password); err != nil {
					log.Error(err)
				} else if user != nil {
					http.SetCookie(w, &http.Cookie{Name: "auth_redirect", MaxAge: -1, Path: "/"})
					http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
					// Redirect to where we were!
					return
				}
			}
		}
	} else {
		// Already authenticated, lets redirect
		http.SetCookie(w, &http.Cookie{Name: "auth_redirect", MaxAge: -1, Path: "/"})
		http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
	}

	// fallback to static pages
	fileserver.ServeHTTP(w, r)
}
