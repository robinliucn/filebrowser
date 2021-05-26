package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/tomasen/realip"

	"github.com/robinliucn/filebrowser/v2/rules"
	"github.com/robinliucn/filebrowser/v2/runner"
	"github.com/robinliucn/filebrowser/v2/settings"
	"github.com/robinliucn/filebrowser/v2/storage"
	"github.com/robinliucn/filebrowser/v2/users"
)

type handleFunc func(w http.ResponseWriter, r *http.Request, d *data) (int, error)

type data struct {
	*runner.Runner
	settings *settings.Settings
	server   *settings.Server
	store    *storage.Storage
	user     *users.User
	raw      interface{}
}

// Check implements rules.Checker.
func (d *data) Check(path string) bool {
	if d.user.HideDotfiles && rules.MatchHidden(path) {
		return false
	}

	allow := true
	for _, rule := range d.settings.Rules {
		if rule.Matches(path) {
			allow = rule.Allow
		}
	}

	for _, rule := range d.user.Rules {
		if rule.Matches(path) {
			allow = rule.Allow
		}
	}

	return allow
}

func handle(fn handleFunc, prefix string, store *storage.Storage, server *settings.Server) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		settings, err := store.Settings.Get()
		if err != nil {
			log.Fatalf("ERROR: couldn't get settings: %v\n", err)
			return
		}

		status, err := fn(w, r, &data{
			Runner:   &runner.Runner{Enabled: server.EnableExec, Settings: settings},
			store:    store,
			settings: settings,
			server:   server,
		})

		if status != 0 {
			txt := http.StatusText(status)
			http.Error(w, strconv.Itoa(status)+" "+txt, status)
		}

		if status >= 400 || err != nil {
			clientIP := realip.FromRequest(r)
			log.Printf("%s: %v %s %v", r.URL.Path, status, clientIP, err)
		}
	})

	return stripPrefix(prefix, handler)
}
