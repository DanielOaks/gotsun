// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/DanielOaks/gotsun/lib"
	"github.com/docopt/docopt-go"
	"github.com/oragono/oragono/mkcerts"
)

func main() {
	version := lib.SemVer
	usage := `gotsun.
Usage:
	gotsun initdb [--conf <filename>]
	gotsun upgradedb [--conf <filename>]
	gotsun mkcerts [--conf <filename>]
	gotsun run [--conf <filename>]
	gotsun -h | --help
	gotsun --version
Options:
	--conf <filename>  Configuration file to use [default: diary.yaml].
	-h --help          Show this screen.
	--version          Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, version, false)

	configfile := arguments["--conf"].(string)
	config, err := lib.LoadConfig(configfile)
	if err != nil {
		log.Fatal("Config file did not load successfully:", err.Error())
	}
	if arguments["initdb"].(bool) {
		lib.InitDB(config.Database.Path)
		if !arguments["--quiet"].(bool) {
			log.Println("database initialized: ", config.Database.Path)
		}
	} else if arguments["upgradedb"].(bool) {
		lib.UpgradeDB(config.Database.Path)
		if !arguments["--quiet"].(bool) {
			log.Println("database upgraded: ", config.Database.Path)
		}
	} else if arguments["mkcerts"].(bool) {
		if !arguments["--quiet"].(bool) {
			log.Println("making self-signed certificates")
		}

		for name, conf := range config.TLSListenersInfo {
			if !arguments["--quiet"].(bool) {
				log.Printf(" making cert for %s listener\n", name)
			}
			err := mkcerts.CreateCert("GotsunDiary", "localhost", conf.Cert, conf.Key)
			if err == nil {
				if !arguments["--quiet"].(bool) {
					log.Printf("  Certificate created at %s : %s\n", conf.Cert, conf.Key)
				}
			} else {
				log.Fatal("  Could not create certificate:", err.Error())
			}
		}
	} else if arguments["run"].(bool) {
		rand.Seed(time.Now().UTC().UnixNano())
		log.Println(fmt.Sprintf("GotsunDiary v%s starting", lib.SemVer))

		// warning if running a non-final version
		if strings.Contains(lib.SemVer, "unreleased") {
			log.Println("You are currently running an unreleased beta version of GotsunDiary that may be unstable and could corrupt your database.")
		}

		server, err := lib.NewServer(config)
		if err != nil {
			log.Fatalln(fmt.Sprintf("Could not load server: %s", err.Error()))
			return
		}

		log.Println("Server running")
		defer log.Println(fmt.Sprintf("GotsunDiary v%s exiting", lib.SemVer))

		server.Run()
	}
}
