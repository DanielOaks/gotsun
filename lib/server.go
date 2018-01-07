// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package lib

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/oragono/oragono/irc/passwd"
	"github.com/tidwall/buntdb"
)

// Server handles incoming queries.
type Server struct {
	Config    *Config
	passwords *passwd.SaltedManager
	templates *pongo2.TemplateSet
	store     *buntdb.DB
}

// NewServer returns a new server using the given config.
func NewServer(config *Config) (*Server, error) {
	server := Server{
		Config: config,
	}
	err := server.loadDatabase(config.Database.Path)
	if err != nil {
		return nil, err
	}
	tmpl, err := loadTemplates(config.Templates)
	if err != nil {
		return nil, err
	}
	server.templates = tmpl

	return &server, nil
}

// loadDatabase, as expected, opens up our database.
func (server *Server) loadDatabase(databasePath string) error {
	// open the database and load server state for which it (rather than config)
	// is the source of truth

	fmt.Println("Opening database")
	db, err := OpenDatabase(databasePath)
	if err == nil {
		server.store = db
	} else {
		return fmt.Errorf("Failed to open datastore: %s", err.Error())
	}

	// load password manager
	fmt.Println("Loading passwords")
	err = server.store.View(func(tx *buntdb.Tx) error {
		saltString, err := tx.Get(keySalt)
		if err != nil {
			return fmt.Errorf("Could not retrieve salt string: %s", err.Error())
		}

		salt, err := base64.StdEncoding.DecodeString(saltString)
		if err != nil {
			return err
		}

		pwm := passwd.NewSaltedManager(salt)
		server.passwords = &pwm
		return nil
	})
	if err != nil {
		return fmt.Errorf("Could not load salt: %s", err.Error())
	}

	return nil
}

// indexHandler handles the index page
func (server *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	result, err := server.templates.RenderTemplateFile("base.tmpl", pongo2.Context{
		"title":   "aaaaa",
		"content": "bbbbbb",
	})

	if err != nil {
		fmt.Fprintln(w, "Could not execute template")
		fmt.Println(err.Error())
		return
	}

	fmt.Fprintf(w, result)
}

// loginHandler handles the login page
func (server *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	result, err := server.templates.RenderTemplateFile("login.tmpl", pongo2.Context{
		"title": "looooooogin",
	})

	if err != nil {
		fmt.Fprintln(w, "Could not execute template")
		fmt.Println(err.Error())
		return
	}

	fmt.Fprintf(w, result)
}

// Run starts the server
func (server *Server) Run() {
	// start router
	r := mux.NewRouter()

	// serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(server.Config.StaticFiles))))

	rg := r.Methods("GET").Subrouter()

	rg.HandleFunc("/", server.indexHandler)
	rg.HandleFunc("/login", server.loginHandler)

	// rg := r.Methods("GET").Subrouter()
	// rg.HandleFunc("/info", restInfo)
	// rg.HandleFunc("/status", restStatus)
	// rg.HandleFunc("/xlines", restGetXLines)
	// rg.HandleFunc("/accounts", restGetAccounts)

	// PUT methods
	// rp := r.Methods("POST").Subrouter()
	// rp.HandleFunc("/rehash", restRehash)

	// make waitgroup
	var wg sync.WaitGroup

	// start listeners
	for _, addr := range server.Config.Listeners {
		// mark us as existing
		wg.Add(1)

		go func(address string, tlsConfig *TLSListenConfig) {
			if tlsConfig != nil {
				fmt.Println("Listening with HTTPS on", address)
				http.ListenAndServeTLS(address, tlsConfig.Cert, tlsConfig.Key, r)
			} else {
				fmt.Println("Listening on", address)
				http.ListenAndServe(address, r)
			}

			// mark this listener as done
			wg.Done()
		}(addr, server.Config.TLSListenersInfo[addr])
	}

	// wait for all listeners to be done before exiting
	wg.Wait()
}
