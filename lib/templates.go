// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"fmt"

	"github.com/flosch/pongo2"
)

// loadTemplates returns our templates.
func loadTemplates(path string) (*pongo2.TemplateSet, error) {
	loader, err := pongo2.NewLocalFileSystemLoader(path)
	if err != nil {
		return nil, fmt.Errorf("Could not load templates: %s", err.Error())
	}

	tset := pongo2.NewSet("web", loader)

	return tset, nil
}
