package main

import (
	"fmt"
	"github.com/nmaupu/gonaomi/core"
	"log"
	"net"
	"time"
)

const (
	ADDRESS = "192.168.12.36"
	PORT    = "10703"
)

func main() {
	fmt.Println("Welcome to GoNaomi")

	strAddr := net.JoinHostPort(ADDRESS, PORT)
	fmt.Println("Connecting to", strAddr, "...")
	conn, err := net.Dial("tcp4", strAddr)
	if err != nil || conn == nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	fmt.Println("Connected to", strAddr)

	phase1(conn)
	//phase2(conn)
	//phase3(conn)
}

func phase1(conn net.Conn) {
	core.HOST_SetMode(conn, 0, 1)
	core.SECURITY_SetKeycode(conn)
}

func phase2(conn net.Conn) {
	core.DIMM_UploadFile(conn, "AW-Metal-Slug-6.bin")
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
