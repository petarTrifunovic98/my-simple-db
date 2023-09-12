package ioprovider

import (
	"net"
	"strings"
)

type SocketIOProvider struct {
	hostName   string
	port       string
	protocol   string
	server     net.Listener
	connection net.Conn
}

func NewSocketIOProvider(server net.Listener) *SocketIOProvider {
	// server, _ := net.Listen(protocol, hostName+":"+port)
	connection, _ := server.Accept()

	socketInputProvider := &SocketIOProvider{
		// hostName:   hostName,
		// port:       port,
		// protocol:   protocol,
		server:     server,
		connection: connection,
	}

	return socketInputProvider
}

func (sip *SocketIOProvider) GetInput() (string, error) {
	buffer := make([]byte, 1024)
	mLen, err := sip.connection.Read(buffer)
	input := string(buffer[:mLen])
	input = strings.TrimSpace(input)
	return input, err
}

func (sip *SocketIOProvider) Print(data string) {
	data += "\n"
	sip.connection.Write([]byte(data))
}
