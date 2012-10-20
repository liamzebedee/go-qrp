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

func newTCPConn(network, addr string) (*TCPConn, error) {
	tcpListener, err := net.ListenTCP(network, addr)
	if err != nil {
		return nil, err
	}
	
	packetQueue := nil
	tcpConn := TCPConn{tcpListener, packetQueue}

	return &udpConn, nil
}

type TCPConn struct {
	tcpListener net.TCPListener
	packetQueue int //*Queue
}

func (conn *TCPConn) Listen() () {
	for {
		conn, err := conn.tcpListener.Accept()
		if err != nil {
			// handle error
			continue
		}
		go func(conn) {
			// Transfer packets into packetQueue
		} ()
	}
}

type tcpPacket struct {
	addr net.Addr
	buffer []byte
	size int
	err error
}

func (conn *TCPConn) ReadNextPacket() (buffer []byte, read int, addr net.Addr, err error) {
	/*
	packet := conn.packetQueue.Pop()
	return packet.buffer, packet.size, packet.addr, packet.err
	*/
	return buffer, read, addr, err
}

// Creates a node using UDP, returning an error if failure
func CreateNodeTCP(net, addr string) (*Node, error) {
	tcpConn, err := newTCPConn(net, addr)
	if err != nil {
		return nil, err
	}

	return CreateNode(udpConn)
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