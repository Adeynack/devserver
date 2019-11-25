package main

import "github.com/adeynack/devserver/pkg/devserver"

func main() {
	devserver.Start(devserver.Configuration{
		ListenAddress:         "localhost:3001",
		HttpDevConfigurations: []devserver.HttpDevConfiguration{
			{
				DestinationAddress: "localhost:3000",
				MountURLPrefix:     "",
			},
		},
	})
}
