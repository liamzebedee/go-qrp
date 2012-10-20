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
	packetQueue chan packet
	activeConnections map[string] net.Conn
	open bool
}

func newTCPModule(network, addr string) (*TCPModule, error) {
	tcpListener, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	
	packetQueue := make(chan packet)
	activeConnections := make(map[string] net.Conn)
	tcpModule := TCPModule{ tcpListener, packetQueue, activeConnections, true }

	return &tcpModule, nil
}

// Listens for connections, and instantiates a new goroutine for each Accept
func (tcpModule *TCPModule) listen() () {
	for {
		conn, err := tcpModule.listener.Accept()
		if err != nil {
			// Not much can be done
			fmt.Errorf("qrp:", "connection not accepted - ", err.Error())
			continue
		}
		fmt.Printf("New connection on %s\n", conn.RemoteAddr().String())
		go tcpModule.handleConnection(conn)
	}
}

func (tcpModule *TCPModule) handleConnection(conn net.Conn) {
	// Store connection
	tcpModule.activeConnections[conn.RemoteAddr().String()] = conn
	reader := bufio.NewReader(conn)
	
	// Read next message
	for {
		fmt.Printf("Packet received on %s\n", conn.RemoteAddr().String())
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
		p := packet{ buffer, read, conn.RemoteAddr(), nil }
		tcpModule.packetQueue <- p
	}
	
	conn.Close()
}

func (tcpModule *TCPModule) ReadNextPacket() (packet) {
	packet := <-tcpModule.packetQueue
	
	return packet
}

func (tcpModule *TCPModule) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if(!tcpModule.open) {
		return 0, nil
	}
	
	tcpConn := tcpModule.activeConnections[addr.String()]
	
	if tcpConn == nil {
		tcpConn, err := net.Dial(addr.Network(), addr.String())
		fmt.Printf("Connecting to Addr: %s[%s]\n", addr.Network(), addr.String())
		fmt.Printf("Connecting from Addr: %s[%s]\n", tcpConn.LocalAddr().Network(), tcpConn.LocalAddr().String())
		if err != nil {
			return 0, err
		}
		
		tcpModule.activeConnections[addr.String()] = tcpConn
	}
	
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
func CreateNodeTCP(net, addr string) (*Node, error) {
	tcpModule, err := newTCPModule(net, addr)
	if err != nil {
		return nil, err
	}
	
	go tcpModule.listen()

	return CreateNode(tcpModule)
}

// Calls a procedure on a node using the TCP protocol. 
// See Node.Call
func (node *Node) CallTCP(procedure string, addrString string, args interface{}, reply interface{}, timeout int) (err error) {
	addr, err := net.ResolveTCPAddr("tcp", addrString)
	if err != nil {
		return err
	}

	return node.Call(procedure, addr, args, reply, timeout)
}