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

// Implementation details:
// - Frames are delimited by the NUL byte (0x00)
// - There is always a single goroutine running - the connection listener. 
// - The connection listener spawns goroutines for every connection, where packets are read and sent to the packetQueue
// - ReadNextPacket returns packets from the packetQueue

import (
	"net"
	"bufio"
	"io"
	"fmt"
)

type TCPModule struct {
	listener net.Listener
	packetQueue chan tcpPacket
	activeConnections map[string] net.Conn
	open bool
}

func newTCPModule(network, addr string) (*TCPModule, error) {
	tcpListener, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	
	packetQueue := make(chan tcpPacket)
	activeConns := make(map[string] net.Conn)
	tcpModule := TCPModule{ tcpListener, packetQueue, activeConns, true }

	return &tcpModule, nil
}

// Listens for connections, and instantiates a new goroutine for each Accept
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
	// Store connection
	tcpModule.activeConnections[conn.RemoteAddr().String()] = conn
	reader := bufio.NewReader(conn)
	
	// Read next message
	for {		
		// Read from stream until NUL byte
		buffer, err := reader.ReadBytes(0x00)
		read := len(buffer)
		if err == io.EOF && read == 0 {
			continue
		} else if err != nil {
			// Close conn, bad packet format
			fmt.Errorf("qrp:", "Closing connection of", conn.RemoteAddr().String(), "-", "Bad packet format")
			break
		}
		
		// Send to packet queue
		packet := tcpPacket{ conn.RemoteAddr(), buffer, read }
		tcpModule.packetQueue <- packet
	}
	
	conn.Close()
}

type tcpPacket struct {
	addr net.Addr
	buffer []byte
	read int
}

func (tcpModule *TCPModule) ReadNextPacket() (buffer []byte, read int, addr net.Addr, err error) {
	packet := <-tcpModule.packetQueue
	
	return packet.buffer, packet.read, packet.addr, nil
}

func (tcpModule *TCPModule) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if(!tcpModule.open) {
		return 0, nil
	}
	
	tcpConn := tcpModule.activeConnections[addr.String()]
	writer := bufio.NewWriter(tcpConn)
	
	n, err = writer.Write(append(b, 0x00))
	if err != nil {
		return 0, err
	}
	
	return n, err
}

func (tcpModule *TCPModule) Close() error {
	tcpModule.open = false
	
	// Stop accepting new connections
	err := tcpModule.listener.Close()
	if err != nil {
		tcpModule.open = true
		return err
	}
	
	// Close all active connections
	for _, conn := range tcpModule.activeConnections {
		conn.Close()
	}
	
	return nil
}

// Creates a TCP node, returning an error if failure
func CreateNodeTCP(addr string) (*Node, error) {
	tcpModule, err := newTCPModule("tcp", addr)
	if err != nil {
		return nil, err
	}
	
	go tcpModule.listen()

	return CreateNode(tcpModule)
}

// Calls a procedure on a node using the TCP protocol. 
// See Node.Call
func (node *Node) CallTCP(procedure string, addrString string, args interface{}, reply interface{}, timeout int) (err error) {
	addr, err := net.ResolveIPAddr("ip", addrString)
	if err != nil {
		return err
	}

	return node.Call(procedure, addr, args, reply, timeout)
}