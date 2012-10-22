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
	"sync"
	"fmt"
	"bufio"
	"io"
)

// Creates a TCP node, returning an error if failure
func CreateNodeTCP(net, listenAddr, connectAddr string) (*Node, error) {
	tcpModule, err := newTCPModule(net, listenAddr, connectAddr)
	if err != nil {
		return nil, err
	}
	
	go tcpModule.listen()

	return CreateNode(tcpModule)
}

type TCPModule struct {
	routines sync.WaitGroup
	listenRoutine chan bool
	
	listener *net.TCPListener
	connectAddr *net.TCPAddr // Address by which we connect to other nodes
	activeConnections map[string] net.Conn
	closed bool
	
	packetQueue chan packet
}

func newTCPModule(network, listenAddrStr, connectAddrStr string) (*TCPModule, error) {
	listenAddr, err := net.ResolveTCPAddr(network, listenAddrStr)
	if err != nil {
		return nil, err
	}
	
	connectAddr, err := net.ResolveTCPAddr(network, connectAddrStr)
	if err != nil {
		return nil, err
	}
	
	tcpListener, err := net.ListenTCP(network, listenAddr)
	if err != nil {
		return nil, err
	}
	
	tcpModule := TCPModule { }
	tcpModule.listener = tcpListener
	tcpModule.connectAddr = connectAddr
	tcpModule.listenRoutine = make(chan bool)
	tcpModule.packetQueue = make(chan packet)
	tcpModule.activeConnections = make(map[string] net.Conn)
	tcpModule.closed = false
	
	go tcpModule.listen()
	
	return &tcpModule, nil
}

// Listens for connections, and instantiates a new goroutine for each Accept
func (tcpModule *TCPModule) listen() () {
	tcpModule.routines.Add(1)
	defer tcpModule.routines.Done()
	
	select {
		case <-tcpModule.listenRoutine:
			tcpModule.listener.Close()
			return
		default:
			conn, err := tcpModule.listener.Accept()
			if err != nil {
				// Not much can be done
				fmt.Printf("qrp: connection not accepted - %s\n", err.Error())
			} else {
				go tcpModule.handleConnection(conn)
			}
	}
}

func (tcpModule *TCPModule) handleConnection(conn net.Conn) {
	tcpModule.routines.Add(1)
	defer tcpModule.routines.Done()
	
	// Store connection
	tcpModule.activeConnections[conn.RemoteAddr().String()] = conn
	reader := bufio.NewReader(conn)
	
	// Read next message
	for tcpModule.closed {
		// Read from stream until NUL byte
		buffer, err := reader.ReadBytes(0)
		read := len(buffer)
		
		if err == io.EOF && read == 0 {
			// No packet, no cry
			continue
			
		} else if err != nil {
			// Close conn, bad packet format
			fmt.Errorf("qrp: Closing connection of %s - Bad packet format\n")
			return
			
		} else {
			fmt.Printf("Packet received from %s\n", conn.RemoteAddr().String())
		
			// Send to packet queue
			p := packet{ buffer, read, conn.RemoteAddr(), nil }
			tcpModule.packetQueue <- p
		}
			
	}
	
	conn.Close()
}

func (tcpModule *TCPModule) ReadNextPacket() (packet) {
	if(tcpModule.closed) {
		return packet{}
	}
	return <-tcpModule.packetQueue
}

func (tcpModule *TCPModule) Close() error {
	// Signal listen goroutine to close
	tcpModule.listenRoutine <- true
	
	// Signal all connection goroutines to close
	tcpModule.closed = true
	
	// Wait for goroutines to close
	tcpModule.routines.Wait()
	
	return nil
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

func (tcpModule *TCPModule) WriteTo(b []byte, addr net.Addr) (int, error) {
	if tcpModule.closed {
		return 0, nil
	}
	
	tcpConn := tcpModule.activeConnections[addr.String()]
	
	// First time we are connecting
	if tcpConn == nil {
		addrConnectTo, err := net.ResolveTCPAddr("tcp", addr.String())
		if err != nil {
			return 0, err
		}
		
		fmt.Printf("%s connecting to %s\n", tcpModule.connectAddr.String(), addrConnectTo.String())
		
		tcpConn, err := net.DialTCP(addr.Network(), tcpModule.connectAddr, addrConnectTo)
		if err != nil {
			fmt.Errorf("qrp:", "error writing to connection -", err.Error())
			return 0, err
		}
		
		go tcpModule.handleConnection(tcpConn)
	}
	writer := bufio.NewWriter(tcpConn)
	
	n, err := writer.Write(append(b, 0x00))
	if err != nil {
		return 0, err
	}
	
	return n, nil
}