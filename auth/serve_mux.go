package auth

import (
	"net/http"
	"sort"

	"github.com/rollerderby/go/json"
)

type serveMux struct {
	mux      *http.ServeMux
	Files    http.Handler
	handlers []*handler
}

func NewServeMux() *serveMux {
	ServeMux = &serveMux{
		mux:   http.NewServeMux(),
		Files: http.FileServer(http.Dir("html")),
	}

	return ServeMux
}

func (sm *serveMux) Handle(display, path string, handler http.Handler, requiredGroups []string) {
	h := Handler(handler, display, path, requiredGroups)
	sm.handlers = append(sm.handlers, h)
	sm.mux.Handle(path, h)
}

func (sm *serveMux) HandleFunc(display, path string, handler func(http.ResponseWriter, *http.Request), requiredGroups []string) {
	h := HandlerFunc(handler, display, path, requiredGroups)
	sm.handlers = append(sm.handlers, h)
	sm.mux.Handle(path, h)
}

func (sm *serveMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	return sm.mux.Handler(r)
}

func (sm *serveMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sm.mux.ServeHTTP(w, r)
}

type MenuItem struct {
	Display  string
	Path     string
	Priority int
}

type menuItems []*MenuItem

func (ml menuItems) Len() int      { return len(ml) }
func (ml menuItems) Swap(i, j int) { ml[i], ml[j] = ml[j], ml[i] }
func (ml menuItems) JSON() json.Value {
	var arr json.Array
	for _, item := range ml {
		obj := make(json.Object)
		obj["Display"] = json.NewString(item.Display)
		obj["Path"] = json.NewString(item.Path)
		obj["Priority"] = json.NewNumber(int64(item.Priority))
		arr = append(arr, obj)
	}
	return arr
}

type byPriorityDisplay struct{ menuItems }

func (ml byPriorityDisplay) Less(i, j int) bool {
	if ml.menuItems[i].Priority != ml.menuItems[j].Priority {
		return ml.menuItems[i].Priority < ml.menuItems[j].Priority
	}

	return ml.menuItems[i].Display < ml.menuItems[j].Display
}

func MenuItems(u *User) menuItems {
	var ret []*MenuItem

	ret = append(ret, &MenuItem{Display: "Home", Path: "/", Priority: 0})
	if u != nil {
		ret = append(ret, &MenuItem{Display: "Logout", Path: "/auth/logout/", Priority: 1})
	} else {
		ret = append(ret, &MenuItem{Display: "Login", Path: "/auth/", Priority: 1})
	}
	for _, h := range ServeMux.handlers {
		if h.display != "" && h.path != "" && h.IsUserAllowed(u) {
			ret = append(ret, &MenuItem{Display: h.display, Path: h.path, Priority: 2})
		}
	}

	sort.Sort(byPriorityDisplay{ret})
	return ret
}
