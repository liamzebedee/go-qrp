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

// UDP connection implementation

import (
	"net"
	"fmt"
	"sync"
)

type UDPNode struct {
	Node
	conn net.UDPConn
	mtu uint32
	
	stopServing chan bool
	routines sync.WaitGroup
}

// Creates a node that communicates over UDP. mtu specifies the maximum transmission unit,
// the maximum size of a packet that can be transmitted. 
func CreateNodeUDP(network, addr string, mtu uint32) (*UDPNode, error) {
	// Connection	
	udpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		return nil, err
	}

	node := CreateNode()
	udpNode := UDPNode{ Node: node, conn: *conn, mtu: mtu }
	udpNode.stopServing = make(chan bool)
	udpNode.Node.Connection = &udpNode
	
	return &udpNode, nil
}

// Listens for queries and replies, serving procedures registered by Register
// Returns an error if there was a failure serving or we are already serving
func (node *UDPNode) ListenAndServe() (error) {
	node.routines.Add(1)
	defer node.conn.Close()
	defer node.routines.Done()
	
	packets := make(chan packet)
	receiverSignaller := make(chan bool) // To signal the receiver goroutine to end
	
	// A seperate receiving goroutine is required so we don't block on the readNextPacket function
	go func() {
		/*node.routines.Add(1)
		defer node.routines.Done()*/
		for {
			select {
				case <-receiverSignaller:
					return
				default:
					// TODO: Make readNextPacket time out so this goroutine can exit
					//		 Once this works uncomment the above node.routines code
					packets <- node.readNextPacket() // blocking
			}
		}
	} ()
	
	/*if node.routines != nil {
		return errors.New("Already serving")
	} : TODO */ 
	
	for {
		select {
			case <-node.stopServing:
				// Signal to stop server
				fmt.Println("qrp:", "Closing server")				
				receiverSignaller <- true
				return nil
				
			case packet := <-packets:
				if packet.err != nil {
					fmt.Errorf("qrp:", "Error reading from connection - %s\n", packet.err.Error())
					continue
				}
	
				// If we read a packet
				if packet.read > 0 {
					// Process packet
					go func() {
						node.routines.Add(1)
						defer node.routines.Done()
						
						err := node.processPacket(packet.buffer, packet.read, packet.addr)
						if err != nil {
							fmt.Errorf("qrp:", "Error processing packet - %s\n", packet.err.Error())
						}
					} ()
				}
		}
	}
	
	return nil
}

// Reads the next packet from the connection and returns the buffer, bytes read and any errors
// This is designed so we can work with multiple protocols, without having to specify buffer sizes	
func (node *UDPNode) readNextPacket() (packet) {
	buffer := make([]byte, node.mtu)

	// Read a packet into the buffer
	read, addr, err := node.conn.ReadFrom(buffer)

	return packet{ buffer, read, addr, err }
}

func (node *UDPNode) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return node.conn.WriteTo(b, addr)
}

// Calls a procedure on a node using the UDP protocol
// see Node.Call
func (node *UDPNode) CallUDP(procedure string, addrString string, args interface{}, reply interface{}, timeout int) (err error) {
	addr, err := net.ResolveUDPAddr("udp", addrString)
	if err != nil {
		return err
	}

	return node.Call(procedure, addr, args, reply, timeout)
}

func (node *UDPNode) Stop() (error) {
	// Signal server to stop
	node.stopServing <- true
	
	node.routines.Wait()
	
	return nil
}