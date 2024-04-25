package main

import (
	"github.com/go-courier/courier"
	"github.com/utilsgo/httptransport/testdata/server/cmd/app/routes"

	"github.com/utilsgo/httptransport"
)

func main() {
	ht := &httptransport.HttpTransport{
		Port: 8080,
	}
	ht.SetDefaults()

	courier.Run(routes.RootRouter, ht)
}
