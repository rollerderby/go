package auth

import "net/http"

type handler struct {
	handler http.Handler
	display string
	path    string
	groups  []string
}

func Handler(h http.Handler, display, path string, g []string) *handler {
	return &handler{
		handler: h,
		display: display,
		path:    path,
		groups:  g,
	}
}

func HandlerFunc(f func(http.ResponseWriter, *http.Request), display, path string, g []string) *handler {
	return Handler(http.HandlerFunc(f), display, path, g)
}

func (h *handler) IsAllowed(r *http.Request) (*UserHelper, bool) {
	if len(h.groups) == 0 {
		return nil, true
	}
	user := CheckAuth(r)
	return user, h.IsUserAllowed(user)
}

func (h *handler) IsUserAllowed(u *UserHelper) bool {
	if len(h.groups) == 0 {
		return true
	}
	if u == nil {
		return false
	}

	return u.HasGroup(h.groups...)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, allowed := h.IsAllowed(r)
	if !allowed {
		if user == nil {
			http.SetCookie(w, &http.Cookie{Name: "auth_redirect", Value: r.URL.String(), MaxAge: 0, Path: "/"})
			http.Redirect(w, r, "/auth/", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
		return
	}
	h.handler.ServeHTTP(w, r)
}
