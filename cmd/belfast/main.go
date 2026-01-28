package main

import "github.com/ggmolly/belfast/internal/entrypoint"

func main() {
	entrypoint.Run(entrypoint.Options{
		CommandName:   "belfast",
		Description:   "Azur Lane server emulator",
		DefaultConfig: "server.toml",
	})
}
