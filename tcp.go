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
// - The connection listener spawns a new goroutine for each connection
// - Upon shutdown, all existing packets being processed are waited for completion, to prevent corruption

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type TCPNode struct {
	// Net
	listenAddr         *net.TCPAddr
	connectAddr        *net.TCPAddr
	activeConnections  map[string]*tcpConn
	_activeConnections *sync.Mutex

	// Sync
	routines             sync.WaitGroup
	listenRoutine_signal chan bool

	// QRP
	Node
}

type tcpConn struct {
	conn           *net.TCPConn
	routine_signal chan bool
	routines       sync.WaitGroup
}

func CreateNodeTCP(network, listenAddrStr, connectAddrStr string) (*TCPNode, error) {
	// Setup listener
	listenAddr, err := net.ResolveTCPAddr(network, listenAddrStr)
	if err != nil {
		return nil, err
	}

	connectAddr, err := net.ResolveTCPAddr(network, connectAddrStr)
	if err != nil {
		return nil, err
	}

	// Setup node
	node := CreateNode()

	// Create TCP node
	tcpNode := TCPNode{
		listenAddr:         listenAddr,
		connectAddr:        connectAddr,
		activeConnections:  make(map[string]*tcpConn),
		_activeConnections: new(sync.Mutex),

		listenRoutine_signal: make(chan bool, 1),

		Node: node,
	}
	node.Connection = &tcpNode

	return &tcpNode, nil
}

func (node *TCPNode) ListenAndServe() error {
	// Sync
	node.routines.Add(1)
	defer node.routines.Done()

	// Setup listener
	listener, err := net.ListenTCP(node.listenAddr.Network(), node.listenAddr)
	if err != nil {
		return err
	}
	listener.SetDeadline(time.Now().Add(2 * time.Second))
	defer listener.Close()

	for {
		select {
		case <-node.listenRoutine_signal:
			return nil
		default:
			listener.SetDeadline(time.Now().Add(2 * time.Second))
			conn, err := listener.AcceptTCP()
			if err != nil {
				// Ain't nobody got time for that
				fmt.Printf("Ain't nobody got time for that")
				continue
			}
			go node.handleConnection(conn)
		}
	}

	return nil
}

// Closes a connection to a specific node
func (node *TCPNode) CloseConn(addrStr string) {
	node._activeConnections.Lock()
	node.activeConnections[addrStr].routine_signal <- true
	node._activeConnections.Unlock()
}

func (node *TCPNode) handleConnection(conn *net.TCPConn) {
	node.routines.Add(1)
	defer node.routines.Done()
	defer conn.Close()

	node._activeConnections.Lock()
	tcpConn := tcpConn{
		conn:           conn,
		routine_signal: make(chan bool),
	}
	node.activeConnections[conn.RemoteAddr().String()] = &tcpConn
	node._activeConnections.Unlock()

	// Setup buffered reader
	reader := bufio.NewReader(conn)

	for {
		select {
		case <-tcpConn.routine_signal:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))

			// Frames are delimited by a NUL byte, so we read up until this
			data, err := reader.ReadBytes(0x00)

			if err == io.EOF {
				continue

			} else if err != nil {
				// Bad packet, doesn't end in NUL
				fmt.Printf("qrp: Closing connection from %s - bad packet format", conn.RemoteAddr().String())
				return

			} else {
				// Process packet
				go func() {
					tcpConn.routines.Add(1)
					defer tcpConn.routines.Done()
					node.Node.processPacket(data, len(data), conn.RemoteAddr())
				}()

			}
		}
	}

	tcpConn.routines.Wait()
}

func (node *TCPNode) Stop() error {
	node.listenRoutine_signal <- true
	fmt.Println("qrp: Signalled listen routine to close")

	fmt.Println("qrp: Signalling all active connections to close")
	node._activeConnections.Lock()
	for _, conn := range node.activeConnections {
		conn.routine_signal <- true
	}
	node._activeConnections.Unlock()

	// Now, we wait
	fmt.Println("qrp: Waiting for everything to finish up")
	node.routines.Wait()

	return nil
}
