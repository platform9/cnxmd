package cnxmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	ConnectionClosedErr = "use of closed network connection"
	ConnectionResetErr  = "connection reset by peer"
	MaxTeardownTimeInSeconds = 35
)

//------------------------------------------------------------------------------

// Start a client proxy service that listens on bindAddr:listenPort.
// When a client connects, the proxy connects to destFqdn:destPort,
// sends it a CNXMD header containing the specified key-value pairs,
// then pipes the bytes from the incoming connection.
func ServeClientProxy(
	bindAddr string,
	listenPort int,
	destFqdn string,
	destPort int,
	kv map[string]string,
) {
	addr := fmt.Sprintf("%s:%d", bindAddr, listenPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	log.Printf("listening on %s", listener.Addr().String())
	for {
		cnx, err := listener.Accept()
		if err != nil {
			log.Printf("warning: failed to accept: %s", err)
		}
		cnxId := RandomString(8) 
		log.Printf("[%s] accepted local connection", cnxId)
		go handleConnection(cnx, cnxId, destFqdn, destPort, kv)
	}
}

//------------------------------------------------------------------------------

func handleConnection(
	cnx net.Conn,
	cnxId string,
	destFqdn string,
	destPort int,
	kv map[string]string,
) {
	dest := fmt.Sprintf("%s:%d", destFqdn, destPort)
	remoteCnx, err := net.Dial("tcp", dest)
	if err != nil {
		log.Printf("[%s] failed to dial %s: %s", cnxId, dest, err)
		return
	}
	defer func() {
		err := remoteCnx.Close()
		if err != nil {
			log.Printf("[%s] failed to close remote connection: %s",
				cnxId, err)
		}
	}()
	log.Printf("[%s] connected to %s", cnxId, dest)
	header := HeadLine + "\n"
	for key, value := range kv {
		header = header + key + "=" + value + "\n"
	}
	header = header + "\n"
	written, err := remoteCnx.Write([]byte(header))
	if err != nil {
		log.Printf("[%s] failed to write header: %s", cnxId, err)
		return
	}
	if written != len(header) {
		log.Printf("[%s] failed to write full header: only %d bytes instead of %d",
			cnxId, written, len(header))
		return
	}
	Copycat(cnx.(*net.TCPConn), remoteCnx.(*net.TCPConn), cnxId)
}

//------------------------------------------------------------------------------

func Copycat(client *net.TCPConn, server *net.TCPConn, cnxId string) {
	log.Printf("[%s] Initiating copy between %s and %s", cnxId,
		client.RemoteAddr().String(), server.RemoteAddr().String())

	doCopy := func(s, c *net.TCPConn, cancel chan<- string) {
		numWritten, err := io.Copy(s, c)
		reason := "EOF"
		if err != nil {
			reason = err.Error()
		}
		log.Printf("[%s] Copied %d bytes from %s to %s, finished because: %s",
			cnxId, numWritten, c.RemoteAddr().String(),
			s.RemoteAddr().String(),
			reason)
		if err != nil && !strings.Contains(err.Error(),
			ConnectionClosedErr) && !strings.Contains(err.Error(),
				ConnectionResetErr) {
			log.Printf("[%s] Failed copying connection data: %v",
				cnxId, err)
		}
		log.Printf("[%s] Copy finished for %s -> %s", cnxId,
			c.RemoteAddr().String(), s.RemoteAddr().String())
		err = s.CloseWrite() // propagate EOF signal to destination
		if err != nil {
			log.Printf("[%s] warning: failed to CloseWrite() %s -> %s : %s --ok",
				cnxId, c.RemoteAddr().String(), s.RemoteAddr().String(), err)
		}
		cancel <- c.RemoteAddr().String()
	}

	cancel := make(chan string, 2)
	go doCopy(server, client, cancel)
	go doCopy(client, server, cancel)

	closedSrc := <- cancel
	log.Printf("[%s] 1st source to close: %s", cnxId, closedSrc)
	timer := time.NewTimer(MaxTeardownTimeInSeconds * time.Second)
	select {
	case closedSrc = <-cancel:
		log.Printf("[%s] 2nd source to close: %s (all done)",
			cnxId, closedSrc)
		timer.Stop()
	case <- timer.C:
		log.Printf("[%s] timed out waiting for 2nd source to close",
			cnxId)
	}
}
