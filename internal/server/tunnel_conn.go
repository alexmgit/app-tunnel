package server

import (
	"bufio"
	"net"
)

type TunnelConn struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

func NewTunnelConn(conn net.Conn) *TunnelConn {
	return &TunnelConn{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

func NewTunnelConnWithReader(conn net.Conn, reader *bufio.Reader) *TunnelConn {
	return &TunnelConn{
		Conn:   conn,
		Reader: reader,
	}
}
