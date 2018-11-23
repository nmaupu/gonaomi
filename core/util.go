package core

import (
	"fmt"
	bp "github.com/roman-kachanovsky/go-binary-pack/binary-pack"
	"io"
	"log"
	"net"
	"os"
)

// Reads n bytes from a given socket
func ReadSocket(conn net.Conn, n int) (string, error) {
	buf := make([]byte, n)

	nb, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", buf[:nb]), nil
}

func HOST_SetMode(conn net.Conn, v_and, v_or int) string {
	f := []string{"I", "I"}
	v := []interface{}{0x07000004, ((v_and << 8) | v_or)}

	bp := new(bp.BinaryPack)
	data, err := bp.Pack(f, v)

	log.Println("Writing", data)
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Finished")

	ret, err := ReadSocket(conn, 0x8)
	if err != nil {
		log.Fatalln(err)
	}

	return ret
}

func SECURITY_SetKeycode(conn net.Conn) {
	f := []string{"I", "I", "I", "I", "I", "I", "I", "I", "I"}
	v := []interface{}{0x7F000008, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	bp := new(bp.BinaryPack)
	data, err := bp.Pack(f, v)

	log.Println("Writing", data)
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Finished")
}

func DIMM_UploadFile(conn net.Conn, filename string) {
	addr := uint32(0)
	crc := uint32(0)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	for {
		data := make([]byte, 0x8000)
		n, err := file.ReadAt(data, int64(addr))
		if err == io.EOF {
			break
		}

		DIMM_Upload(conn, addr, data[:n], 0)

		crc = CRC32(crc, data[:n])
		addr += uint32(n)
	}

	DIMM_Upload(conn, addr, []byte("12345678"), 1)

	crc = ^crc
	fmt.Printf("CRC to send: %x\n", crc)
	DIMM_SetInformation(conn, crc, addr)
}

func DIMM_Upload(conn net.Conn, addr uint32, d []byte, mark int) {
	f := []string{"I", "I", "I", "H"}
	v := []interface{}{0x04800000 | (len(d) + 0xA) | (mark << 16), 0, int(addr), 0}

	bp := new(bp.BinaryPack)
	data, err := bp.Pack(f, v)
	log.Println("pack content=", data)
	data = append(data, d...)

	log.Println("Writing data...")
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Finished")
}

func DIMM_SetInformation(conn net.Conn, crc uint32, length uint32) {
	f := []string{"I", "I", "I", "I"}
	v := []interface{}{0x1900000C, int(crc) & 0xFFFFFFFF, int(length), 0}

	fmt.Printf("Length=%08x\n", length)
	fmt.Printf("CRC=%x\n", crc)
	bp := new(bp.BinaryPack)
	data, err := bp.Pack(f, v)

	log.Println("Writing", data)
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Finished")
}

func HOST_Restart(conn net.Conn) {
	f := []string{"I"}
	v := []interface{}{0x0A000000}

	bp := new(bp.BinaryPack)
	data, err := bp.Pack(f, v)

	log.Println("Writing", data)
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Finished")
}

//func computeCRC32(data []byte, crc uint32) uint32 {
//	c := crc32.ChecksumIEEE(data)
//	return c
//}

func TIME_SetLimit(conn net.Conn, lim int) {
	f := []string{"I", "I"}
	v := []interface{}{0x17000004, lim}

	bp := new(bp.BinaryPack)
	data, _ := bp.Pack(f, v)

	log.Println("Writing", data)
	_, _ = conn.Write(data)
	// Ignoring error
	log.Println("Finished")
}
