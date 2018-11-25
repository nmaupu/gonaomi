package core

import (
	"fmt"
	bp "github.com/roman-kachanovsky/go-binary-pack/binary-pack"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	CONNECT_TIMEOUT    = 5
	READ_WRITE_TIMEOUT = 30
)

type Naomi struct {
	Addr        string
	Connection  net.Conn
	ProgressBar bool
}

func NewNaomi(addr string, port int) Naomi {
	strAddr := net.JoinHostPort(addr, strconv.Itoa(port))
	n := Naomi{Addr: strAddr}

	log.Printf("Connecting to Naomi at %s...\n", n.Addr)
	c, err := net.DialTimeout("tcp4", n.Addr, time.Duration(CONNECT_TIMEOUT*time.Second))
	if err != nil || c == nil {
		log.Panicln(err)
	}

	n.Connection = c
	n.ProgressBar = true

	log.Printf("Connected to %s\n", n.Addr)

	return n
}

func (n Naomi) Close() {
	n.Connection.Close()
}

// Check wether Naomi board is up or down (check its TCP connectivity)
func (n Naomi) IsUp() bool {
	conn, err := net.Dial("tcp4", n.Addr)
	defer conn.Close()
	return err != nil
}

func (n Naomi) SetDeadline() {
	n.Connection.SetDeadline(time.Now().Add(time.Second * READ_WRITE_TIMEOUT))
}

// Reads nb bytes from a given socket
func (n Naomi) ReadSocket(nb int) (string, error) {
	buf := make([]byte, nb)

	n.SetDeadline()
	nbRet, err := n.Connection.Read(buf)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", buf[:nbRet]), nil
}

func (n Naomi) WritePacket(format []string, values []interface{}, additionalData []byte) (int, error) {
	nbWrite := -1

	bp := new(bp.BinaryPack)
	data, err := bp.Pack(format, values)
	if err != nil {
		return nbWrite, err
	}

	//log.Printf("Writing %x...\n", data)

	if additionalData != nil {
		data = append(data, additionalData...)
	}

	n.SetDeadline()
	nbWrite, err = n.Connection.Write(data)
	if err != nil {
		log.Println(err)
	}

	return nbWrite, err
}

func (n Naomi) HOST_SetMode(v_and, v_or int) string {
	f := []string{"I", "I"}
	v := []interface{}{0x07000004, ((v_and << 8) | v_or)}

	_, err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Panicln(err)
	}

	ret, err := n.ReadSocket(0x8)
	if err != nil {
		log.Panicln(err)
	}

	return ret
}

func (n Naomi) SECURITY_SetKeycode() {
	f := []string{"I", "I", "I", "I", "I", "I", "I", "I", "I"}
	v := []interface{}{0x7F000008, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	_, err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Panicln(err)
	}
}

func (n Naomi) DIMM_UploadFile(filename string) {
	addr := uint32(0)
	crc := uint32(0)

	log.Printf("Sending %s...\n", filename)

	file, err := os.Open(filename)
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()

	// Get file size
	fileInfo, _ := file.Stat()

	// Progress bar initialization
	var progress *mpb.Progress
	var bar *mpb.Bar
	var start time.Time
	if n.ProgressBar {
		progress = mpb.New(
			mpb.WithWidth(60),
			mpb.WithFormat("[=>-]"),
			mpb.WithRefreshRate(180*time.Millisecond),
		)
		bar = progress.AddBar(fileInfo.Size(),
			mpb.PrependDecorators(decor.Counters(decor.UnitKiB, "% 6.1f / % 6.1f")),
			mpb.AppendDecorators(decor.Percentage()),
		)
		start = time.Now()
	}

	data := make([]byte, 0x8000)
	currentSent := int64(0)
	for {
		nb, err := file.ReadAt(data, int64(addr))
		if err == io.EOF {
			break
		}

		n.DIMM_Upload(addr, data[:nb], 0)

		crc = CRC32(crc, data[:nb])
		addr += uint32(nb)

		if n.ProgressBar {
			bar.IncrBy(nb, time.Since(start))
		} else {
			currentSent += int64(nb)
			percent := currentSent * 100 / fileInfo.Size()
			log.Printf("Sending ... %d%%\n", percent)
		}
	}

	if n.ProgressBar {
		progress.Wait()
	}

	n.DIMM_Upload(addr, []byte("12345678"), 1)

	crc = ^crc
	n.DIMM_SetInformation(crc, addr)
}

func (n Naomi) DIMM_Upload(addr uint32, d []byte, mark int) {
	f := []string{"I", "I", "I", "H"}
	v := []interface{}{0x04800000 | (len(d) + 0xA) | (mark << 16), 0, int(addr), 0}

	_, err := n.WritePacket(f, v, d)
	if err != nil {
		log.Panicln(err)
	}
}

func (n Naomi) DIMM_SetInformation(crc uint32, length uint32) {
	f := []string{"I", "I", "I", "I"}
	v := []interface{}{0x1900000C, int(crc) & 0xFFFFFFFF, int(length), 0}

	log.Printf("Length=%08x\n", length)
	log.Printf("CRC=%x\n", crc)

	_, err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Panicln(err)
	}
}

func (n Naomi) HOST_Restart() {
	f := []string{"I"}
	v := []interface{}{0x0A000000}

	_, err := n.WritePacket(f, v, nil)
	if err != nil {
		log.Panicln(err)
	}
}

func (n Naomi) TIME_SetLimit(lim int) {
	f := []string{"I", "I"}
	v := []interface{}{0x17000004, lim}

	// Writing packet ignoring errors
	n.WritePacket(f, v, nil)
}

func (n Naomi) NETFIRM_GetInformation() string {
	f := []string{"I"}
	v := []interface{}{0x1e000000}

	n.WritePacket(f, v, nil)
	ret, err := n.ReadSocket(0x404)
	if err != nil {
		log.Panicln(err)
	}

	return string(ret)
}

func (n Naomi) SendSingleFile(filename string) {
	n.HOST_SetMode(0, 1)
	n.SECURITY_SetKeycode()

	n.DIMM_UploadFile(filename)
	n.HOST_Restart()

	log.Println("Entering time limit hack loop...")
	for {
		n.TIME_SetLimit(10 * 60 * 1000)
		time.Sleep(5000 * time.Millisecond)
	}
}
