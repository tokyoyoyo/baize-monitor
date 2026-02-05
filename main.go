package main

import (
	"baize-monitor/internal/server/wire"
	"baize-monitor/pkg/constants"
)

func main() {
	server, err := wire.InitializeServer(constants.ServerConfigPath)
	if err != nil {
		panic(err)
	}

	if err := server.Start(); err != nil {
		panic(err)
	}
}
