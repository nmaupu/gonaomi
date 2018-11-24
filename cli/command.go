package cli

import (
	"flag"
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/nmaupu/gonaomi/core"
	"log"
	"os"
	"syscall"
	"time"
)

const (
	NAOMI_PORT = 10703
)

var (
	ip           *string
	port         *int
	filename     *string
	forceRestart *bool
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
		EnvVar: "NAOMI_PORT",
		Value:  NAOMI_PORT,
	})

	filename = app.String(cli.StringOpt{
		Name:   "f file",
		Desc:   "File to load onto the Naomi board",
		EnvVar: "FILENAME",
	})

	forceRestart = app.Bool(cli.BoolOpt{
		Name:   "force",
		Desc:   "Try to force a restart if the current loaded game does not quit",
		EnvVar: "FORCE_RESTART",
	})

	app.Action = execute
	app.Run(os.Args)
}

func execute() {
	var msgs []string
	if *ip == "" {
		msgs = append(msgs, "IP address must be specified")
	}

	if *filename == "" {
		msgs = append(msgs, "Filename must be specified")
	}

	// Print all parameters' error and exist if need be
	if len(msgs) > 0 {
		fmt.Fprintf(os.Stderr, "The following error(s) occured:\n")
		for _, m := range msgs {
			fmt.Fprintf(os.Stderr, "  - %s\n", m)
		}
		os.Exit(1)
	}
	// End params checking

	fmt.Println("Welcome to GoNaomi")

	naomi := core.NewNaomi(*ip, *port)
	defer naomi.Close()

	if *forceRestart {
		log.Println("Trying to force restart...")
		naomi.HOST_Restart()
	}

	// Phase1 prepare for upload and reboot the board
	phase1(&naomi)

	// Phase2 and 3 upload and reboot the board onto the game
	phase2(&naomi, *filename)
	phase3(&naomi)
}

func phase1(n *core.Naomi) {
	n.HOST_SetMode(0, 1)
	n.SECURITY_SetKeycode()
}

func phase2(n *core.Naomi, filename string) {
	n.DIMM_UploadFile(filename)
	n.HOST_Restart()
}

func phase3(n *core.Naomi) {
	// infinite loop
	log.Println("Entering time limit hack loop...")
	for {
		n.TIME_SetLimit(10 * 60 * 1000)
		time.Sleep(5000 * time.Millisecond)
	}
}
