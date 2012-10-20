/*
   Copyright Liam (liamzebedee) Edwards-Playne 2012

   This file is part of QRP.

   QRP is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   QRP is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with QRP.  If not, see <http://www.gnu.org/licenses/>.
*/

package qrp

// TCP connection implementation

import (
	"net"
)

type TCPModule struct {
	listener net.Listener
	packetQueue chan tcpPacket
}

func newTCPModule(network, addr string) (*TCPModule, error) {
	tcpListener, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	
	packetQueue := make(chan tcpPacket)
	tcpModule := TCPModule{ tcpListener, packetQueue }

	return &tcpModule, nil
}

func (tcpModule *TCPModule) listen() () {
	for {
		conn, err := tcpModule.listener.Accept()
		if err != nil {
			// Not much can be done
			continue
		}
		go tcpModule.handleConnection(conn)
	}
}

func (tcpModule *TCPModule) handleConnection(conn net.Conn) {
	// Read next message
	// Send to packet queue
}

type tcpPacket struct {
	addr net.Addr
	buffer []byte
	size int
	err error
}

func (tcpModule *TCPModule) ReadNextPacket() (buffer []byte, read int, addr net.Addr, err error) {
	
	return buffer, read, addr, err
}

func (tcpModule *TCPModule) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return n, err
}

func (tcpModule *TCPModule) Close() error {
	return nil
}

// Creates a node using UDP, returning an error if failure
func CreateNodeTCP(addr string) (*Node, error) {
	tcpModule, err := newTCPModule("tcp", addr)
	if err != nil {
		return nil, err
	}

	return CreateNode(tcpModule)
}

// Calls a procedure on a node using the UDP protocol
// see Node.Call
func (node *Node) CallTCP(procedure string, addrString string, args interface{}, reply interface{}, timeout int) (err error) {
	addr, err := net.ResolveUDPAddr("ip", addrString)
	if err != nil {
		return err
	}

	return node.Call(procedure, addr, args, reply, timeout)
}