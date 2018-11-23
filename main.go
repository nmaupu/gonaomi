package main

import (
	"github.com/nmaupu/gonaomi/cli"
)

const (
	AppName = "gonaomi"
	AppDesc = "Command line tool to load a game onto a Sega Naomi board written in Go."
)

var (
	AppVersion string
)

func main() {
	if AppVersion == "" {
		AppVersion = "master"
	}

	cli.Process(AppName, AppDesc, AppVersion)
}
