package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/costory-io/costory-terraform/internal/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/costory-io/costory",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New("dev"), opts); err != nil {
		log.Fatal(err.Error())
	}
}
