package base

import (
	"github.com/krane-io/krane/types"
	"net/url"
)

type Driver interface {
	List(parameters url.Values) ([]types.Ship, error)
	Plan(parameters url.Values) ([]types.Plan, error)
	Create(parameters url.Values) (string, error)
	Stop(args map[string]string) error
	Name() string
}
