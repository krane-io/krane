package types

import (
	"net/url"
)

type Driver interface {
	List(parameters url.Values) ([]Ship, error)
	Plan(parameters url.Values) ([]Plan, error)
	Create(parameters url.Values) (string, error)
	Stop(args map[string]string) error
	Name() string
}
