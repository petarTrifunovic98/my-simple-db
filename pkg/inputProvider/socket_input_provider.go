package inputprovider

import (
	"net"
	"strings"
)

type SocketInputProvider struct {
	hostName   string
	port       string
	protocol   string
	server     net.Listener
	connection net.Conn
}

func NewSocketInputProvider(hostName string, port string, protocol string) *SocketInputProvider {
	server, _ := net.Listen(protocol, hostName+":"+port)
	connection, _ := server.Accept()

	socketInputProvider := &SocketInputProvider{
		hostName:   hostName,
		port:       port,
		protocol:   protocol,
		server:     server,
		connection: connection,
	}

	return socketInputProvider
}

func (sip *SocketInputProvider) GetInput() (string, error) {
	buffer := make([]byte, 1024)
	mLen, err := sip.connection.Read(buffer)
	input := string(buffer[:mLen])
	input = strings.TrimSpace(input)
	return input, err
}
