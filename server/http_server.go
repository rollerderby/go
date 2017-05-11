package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/rollerderby/go/auth"
)

func printStartup(port uint16) {
	log.Notice("")
	log.Noticef("CRG Scoreboard and Game System Version %v", version)
	log.Notice("")
	log.Notice("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	log.Notice("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	log.Notice("Open a web browser (either Google Chrome or Mozilla Firefox recommended) to:")
	log.Noticef("http://localhost:%d/", port)
	log.Notice("or try one of these URLs:")
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Notice("Cannot get interfaces:", err)
	} else {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				log.Noticef("Cannot get addresses on %v: %v", i, err)
			} else {
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}

					if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
						continue
					}
					var url string
					if ip.To4() != nil {
						url = fmt.Sprintf("http://%v:%d/", ip, port)
					} else {
						url = fmt.Sprintf("http://[%v]:%d/", ip, port)
					}
					log.Notice(url)
				}
			}
		}
	}
	log.Notice("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	log.Notice("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
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
		return fmt.Errorf("Cannot extract HTML folder: %v", err)
	}

	auth.NewServeMux()
	auth.ServeMux.Handle("", "/", auth.ServeMux.Files, nil)
	auth.ServeMux.Handle("Admin System", "/admin/", auth.ServeMux.Files, []string{"admin"})
	auth.ServeMux.HandleFunc("", "/ws/control", controlHandler, nil)

	c := make(chan error, 1)

	go func() {
		printStartup(port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), httpLog(auth.ServeMux))
		if err != nil {
			c <- err
			signals <- os.Kill
		}
	}()

	// Give HTTP Server 1 second to boot up (or fail)
	select {
	case err := <-c:
		return err
	case <-time.After(time.Second):
		return nil
	}
}
