package main

import (
	"github.com/krane-io/krane/api/client"
)

func main() {

	d := client.NewKraneCli(nil, nil, nil, "nil", "pasas", nil)
	d.ToString()

}
