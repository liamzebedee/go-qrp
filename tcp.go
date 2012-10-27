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
	"time"
)

type TCPNode struct {
	Node
	
	routines sync.WaitGroup
	listener *net.TCPListener
	connectAddr *net.TCPAddr // Address by which we connect to other nodes
	activeConnections map[string] net.Conn
	
	listenRoutine chan bool
	handleConnRoutine chan bool
}

// Creates a TCP node, returning an error if failure
func CreateNodeTCP(network, listenAddrStr, connectAddrStr string) (*TCPNode, error) {
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
	
	node := TCPNode { }
	node.listener = tcpListener
	node.connectAddr = connectAddr
	node.listenRoutine, node.handleConnRoutine = make(chan bool), make(chan bool, 42)
	node.activeConnections = make(map[string] net.Conn)
	node.Node = CreateNode()
	node.Node.Connection = &node
	
	return &node, nil
}

// Listens for connections, and instantiates a new goroutine for each Accept
func (node *TCPNode) ListenAndServe() (error) {
	node.routines.Add(1)
	defer node.listener.Close()
	defer node.routines.Done()
	
	conns := make(chan net.Conn, 1)
	receiveSignaller := make(chan bool)
	
	go func() {
		node.routines.Add(1)
		defer node.routines.Done()
		
		for {
			select {
				case <-receiveSignaller:
					return
				default:
					// TODO: Has to be a better way to do this
					node.listener.SetDeadline(time.Now().Add(2 * time.Second))
					conn, err := node.listener.Accept()
					if err != nil {
						// Not much can be done
						fmt.Printf("qrp: connection not accepted - %s\n", err.Error())
						continue
					} else {
						conns <- conn
					}
			}
		}
	} ()
	
	for {
		select {
			case <-node.listenRoutine:
				receiveSignaller <- true
				return nil
			default:
				go node.handleConnection(<-conns)
		}
	}
	
	return nil
}

func (node *TCPNode) handleConnection(conn net.Conn) {
	node.routines.Add(1)
	defer conn.Close()
	defer node.routines.Done()
	
	// Store connection
	node.activeConnections[conn.RemoteAddr().String()] = conn
	reader := bufio.NewReader(conn)
	
	// TODO: Has to be a better way to do this
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	// Read next message
	for {
		select {
			case <-node.handleConnRoutine:
				return
				
			default:
				// Read from stream until NUL byte
				buffer, err := reader.ReadBytes(0)
				read := len(buffer)
				
				if err == io.EOF && read == 0 {
					// No packet, no cry
					continue
					
				} else if err != nil {
					// Close conn, bad packet format
					fmt.Printf("qrp: Closing connection of %s - Bad packet format\n")
					return
					
				} else {
					fmt.Printf("Packet received from %s\n", conn.RemoteAddr().String())
				
					go node.processPacket(buffer, read, conn.RemoteAddr())
				}
		}
	}
}

func (node *TCPNode) Close() error {
	// Signal listen goroutine to close
	node.listenRoutine <- true
	
	// Signal all connection goroutines to close
	for _, _ = range node.activeConnections {
		node.handleConnRoutine <- true
	}
	
	// Wait for goroutines to close
	node.routines.Wait()
	
	return nil
}

// Calls a procedure on a node using the TCP protocol. 
// See Node.Call
func (node *TCPNode) CallTCP(procedure string, addrString string, args interface{}, reply interface{}, timeout int) (err error) {
	addr, err := net.ResolveTCPAddr("tcp", addrString)
	if err != nil {
		return err
	}

	return node.Call(procedure, addr, args, reply, timeout)
}

func (node *TCPNode) WriteTo(b []byte, addr net.Addr) (int, error) {
	tcpConn := node.activeConnections[addr.String()]
	
	// First time we are connecting
	if tcpConn == nil {
		addrConnectTo, err := net.ResolveTCPAddr("tcp", addr.String())
		if err != nil {
			return 0, err
		}
		
		fmt.Printf("%s connecting to %s\n", node.connectAddr.String(), addrConnectTo.String())
		
		tcpConn, err := net.DialTCP(addr.Network(), node.connectAddr, addrConnectTo)
		if err != nil {
			fmt.Printf("qrp:", "error writing to connection -", err.Error())
			return 0, err
		}
		
		go node.handleConnection(tcpConn)
	}
	
	// TODO: Has to be a better way to do this
	tcpConn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	
	writer := bufio.NewWriter(tcpConn)
	
	n, err := writer.Write(append(b, 0x00))
	if err != nil {
		return 0, err
	}
	
	err = writer.Flush()
	if err != nil {
		return 0, err
	}
	
	return n, nil
}