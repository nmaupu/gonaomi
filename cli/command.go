package cli

import (
	"flag"
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/nmaupu/gonaomi/core"
	"github.com/nmaupu/gonaomi/server"
	"log"
	"os"
	"syscall"
)

const (
	NAOMI_DEFAULT_PORT = 10703
)

var (
	ip   *string
	port *int
)

func Process(appName, appDesc, appVersion string) {
	syscall.Umask(0)
	flag.Set("logtostderr", "true")

	app := cli.App(appName, appDesc)
	app.Version("v version", fmt.Sprintf("%s version %s", appName, appVersion))

	ip = app.String(cli.StringOpt{
		Name:   "a address",
		Desc:   "IP address of the Naomi board",
		EnvVar: "NAOMI_ADDR",
	})

	port = app.Int(cli.IntOpt{
		Name:   "p port",
		Desc:   "Port of the Naomi board",
		EnvVar: "NAOMI_DEFAULT_PORT",
		Value:  NAOMI_DEFAULT_PORT,
	})

	app.Command("send", "Send a single file to the Naomi board", sendMode)
	app.Command("server", "Start in server mode, ready to accept request", serverMode)

	app.Run(os.Args)
}

func sendMode(cmd *cli.Cmd) {
	forceRestart := cmd.Bool(cli.BoolOpt{
		Name:   "force",
		Value:  false,
		Desc:   "Force restart of the Naomi board before proceeding",
		EnvVar: "FORCE_RESTART",
	})

	filename := cmd.String(cli.StringOpt{
		Name:   "f filename",
		Value:  "",
		Desc:   "Filename to load onto the Naomi board",
		EnvVar: "FILENAME",
	})

	cmd.Action = func() {
		naomi := core.NewNaomi(*ip, *port)
		defer naomi.Close()

		if *forceRestart {
			log.Println("Trying to force restart...")
			naomi.HOST_Restart()
		}

		naomi.SendSingleFile(*filename)
	}
}

func serverMode(cmd *cli.Cmd) {
	listenPort := cmd.Int(cli.IntOpt{
		Name:   "listen-port l",
		Value:  8080,
		Desc:   "Server port to listen from",
		EnvVar: "LISTEN_PORT",
	})

	romsPath := cmd.String(cli.StringOpt{
		Name:   "roms-path r",
		Value:  "/tmp",
		Desc:   "Path containing roms",
		EnvVar: "ROMS_PATH",
	})

	cmd.Action = func() {
		server.Start(*listenPort, *ip, *port, *romsPath)
	}
}
