package cnxmd

import (
	"fmt"
	"github.com/platform9/proxylib/pkg/proxylib"
	"log"
	"net"
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
	proxylib.FerryBytes(cnx.(*net.TCPConn), remoteCnx.(*net.TCPConn), cnxId, 0)
}
