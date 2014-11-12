package types

import (
	"net/url"
)

type Driver interface {
	List(parameters url.Values) (Fleet, error)
	Plan(parameters url.Values) ([]Plan, error)
	Create(parameters url.Values) (string, error)
	Destroy(parameters url.Values) (string, error)
	Stop(args map[string]string) error
	Name() string
	FindShip(name string) Ship
	ValidateId(text string) bool
}
