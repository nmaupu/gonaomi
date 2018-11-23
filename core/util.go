package core

import (
	"fmt"
	bp "github.com/roman-kachanovsky/go-binary-pack/binary-pack"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

type Naomi struct {
	Addr       string
	Connection net.Conn
}

func NewNaomi(addr string, port int) Naomi {
	strAddr := net.JoinHostPort(addr, strconv.Itoa(port))
	n := Naomi{Addr: strAddr}

	log.Printf("Connecting to Naomi at %s...\n", n.Addr)
	c, err := net.Dial("tcp4", n.Addr)
	if err != nil || c == nil {
		log.Fatalln(err)
	}

	n.Connection = c

	log.Printf("Connected to %s\n", n.Addr)

	return n
}

func (n Naomi) Close() {
	n.Connection.Close()
}

// Reads nb bytes from a given socket
func (n Naomi) ReadSocket(nb int) (string, error) {
	buf := make([]byte, nb)

	nbRet, err := n.Connection.Read(buf)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", buf[:nbRet]), nil
}

func (n Naomi) WritePacket(format []string, values []interface{}, additionalData []byte) error {
	bp := new(bp.BinaryPack)
	data, err := bp.Pack(format, values)
	if err != nil {
		return err
	}

	log.Printf("Writing %x...\n", data)

	if additionalData != nil {
		data = append(data, additionalData...)
	}

	_, err = n.Connection.Write(data)
	if err != nil {
		return err
	}

	return err
}

func (n Naomi) HOST_SetMode(v_and, v_or int) string {
	f := []string{"I", "I"}
	v := []interface{}{0x07000004, ((v_and << 8) | v_or)}

	err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Fatalln(err)
	}

	ret, err := n.ReadSocket(0x8)
	if err != nil {
		log.Fatalln(err)
	}

	return ret
}

func (n Naomi) SECURITY_SetKeycode() {
	f := []string{"I", "I", "I", "I", "I", "I", "I", "I", "I"}
	v := []interface{}{0x7F000008, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (n Naomi) DIMM_UploadFile(filename string) {
	addr := uint32(0)
	crc := uint32(0)

	log.Printf("Sending %s...\n", filename)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	data := make([]byte, 0x8000)
	for {
		nb, err := file.ReadAt(data, int64(addr))
		if err == io.EOF {
			break
		}

		n.DIMM_Upload(addr, data[:nb], 0)

		crc = CRC32(crc, data[:nb])
		addr += uint32(nb)
	}

	n.DIMM_Upload(addr, []byte("12345678"), 1)

	crc = ^crc
	log.Printf("CRC to send: %x\n", crc)
	n.DIMM_SetInformation(crc, addr)
}

func (n Naomi) DIMM_Upload(addr uint32, d []byte, mark int) {
	f := []string{"I", "I", "I", "H"}
	v := []interface{}{0x04800000 | (len(d) + 0xA) | (mark << 16), 0, int(addr), 0}

	err := n.WritePacket(f, v, d)
	if err != nil {
		log.Fatalln(err)
	}
}

func (n Naomi) DIMM_SetInformation(crc uint32, length uint32) {
	f := []string{"I", "I", "I", "I"}
	v := []interface{}{0x1900000C, int(crc) & 0xFFFFFFFF, int(length), 0}

	log.Printf("Length=%08x\n", length)
	log.Printf("CRC=%x\n", crc)

	err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (n Naomi) HOST_Restart() {
	f := []string{"I"}
	v := []interface{}{0x0A000000}

	err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (n Naomi) TIME_SetLimit(lim int) {
	f := []string{"I", "I"}
	v := []interface{}{0x17000004, lim}

	// Writing packet ignoring errors
	n.WritePacket(f, v, nil)
}
