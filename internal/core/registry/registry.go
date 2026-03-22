package registry

import "fmt"

type Registry interface {
	Register(module string, queries []string) error
	Lookup(query string) (string, error)
	Deregister(module string) error
}

var ErrQueryNotFound = fmt.Errorf("registry: query not found")
