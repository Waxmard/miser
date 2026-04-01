package repository

import "fmt"

var drivers = map[string]func(string) (Repository, error){}

// Register makes a database driver available by name.
func Register(name string, fn func(string) (Repository, error)) {
	drivers[name] = fn
}

// New creates a Repository for the given driver and data source.
func New(driver, dsn string) (Repository, error) {
	fn, ok := drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	return fn(dsn)
}
