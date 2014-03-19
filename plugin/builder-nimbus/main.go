package main

import (
	"github.com/mitchellh/packer/builder/nimbus"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(nimbus.Builder))
	server.Serve()
}
