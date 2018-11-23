package cli

import (
	"flag"
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/nmaupu/gonaomi/core"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	NAOMI_PORT = 10703
)

var (
	ip       *string
	port     *int
	filename *string
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
	/* End params checking */

	fmt.Println("Welcome to GoNaomi")

	strAddr := net.JoinHostPort(*ip, strconv.Itoa(*port))
	fmt.Println("Connecting to", strAddr, "...")
	conn, err := net.Dial("tcp4", strAddr)
	if err != nil || conn == nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	fmt.Println("Connected to", strAddr)

	phase1(conn)
	//phase2(conn, *filename)
	//phase3(conn)
}

func phase1(conn net.Conn) {
	core.HOST_SetMode(conn, 0, 1)
	core.SECURITY_SetKeycode(conn)
}

func phase2(conn net.Conn, filename string) {
	core.DIMM_UploadFile(conn, filename)
	core.HOST_Restart(conn)

}

func phase3(conn net.Conn) {
	//loop
	log.Println("Entering time limit hack loop...")
	for {
		core.TIME_SetLimit(conn, 10*60*1000)
		time.Sleep(5000 * time.Millisecond)
	}
}
