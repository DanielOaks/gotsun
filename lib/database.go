// Copyright (c) 2012-2014 Jeremy Latt
// Copyright (c) 2016 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package lib

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/oragono/oragono/irc/passwd"
	"github.com/tidwall/buntdb"
)

const (
	// 'version' of the database schema
	keySchemaVersion = "db.version"
	// latest schema of the db
	latestDbSchema = "1"
	// key for the primary salt used by the ircd
	keySalt = "crypto.salt"
)

// InitDB creates the database.
func InitDB(path string) {
	// make sure it doesn't already exist
	if _, err := os.Stat(path); err == nil {
		log.Fatal("Database already exists!")
	}

	// prepare kvstore db
	store, err := buntdb.Open(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to open datastore: %s", err.Error()))
	}
	defer store.Close()

	err = store.Update(func(tx *buntdb.Tx) error {
		// set base db salt
		salt, err := passwd.NewSalt()
		encodedSalt := base64.StdEncoding.EncodeToString(salt)
		if err != nil {
			log.Fatal("Could not generate cryptographically-secure salt for the user:", err.Error())
		}
		tx.Set(keySalt, encodedSalt, nil)

		// set schema version
		tx.Set(keySchemaVersion, latestDbSchema, nil)
		return nil
	})

	if err != nil {
		log.Fatal("Could not save datastore:", err.Error())
	}
}

// OpenDatabase returns an existing database, performing a schema version check.
func OpenDatabase(path string) (*buntdb.DB, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Database does not exist, creating it.")
		InitDB(path)
	}

	// open data store
	db, err := buntdb.Open(path)
	if err != nil {
		return nil, err
	}

	// check db version
	err = db.View(func(tx *buntdb.Tx) error {
		version, _ := tx.Get(keySchemaVersion)
		if version != latestDbSchema {
			return fmt.Errorf("Database must be updated. Expected schema v%s, got v%s", latestDbSchema, version)
		}
		return nil
	})

	if err != nil {
		// close the db
		db.Close()
		return nil, err
	}

	return db, nil
}

// UpgradeDB upgrades the datastore to the latest schema.
func UpgradeDB(path string) {
	store, err := buntdb.Open(path)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to open datastore: %s", err.Error()))
	}
	defer store.Close()

	err = store.Update(func(tx *buntdb.Tx) error {
		version, _ := tx.Get(keySchemaVersion)

		fmt.Println("We have no database checks, schema is currently version", version)

		return nil
	})
	if err != nil {
		log.Fatal("Could not update datastore:", err.Error())
	}

	return
}
