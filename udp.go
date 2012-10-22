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
)

func newUDPConn(network, addr string, mtu int32) (*UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	udpConn := UDPConn{*conn, mtu}

	return &udpConn, nil
}

type UDPConn struct {
	net.UDPConn
	mtu int32
}

func (conn *UDPConn) ReadNextPacket() (packet) {
	buffer := make([]byte, conn.mtu)

	// Read a packet into the buffer
	read, addr, err := conn.ReadFrom(buffer)

	return packet{ buffer, read, addr, err }
}

// Creates a node using UDP, returning an error if failure
func CreateNodeUDP(net, addr string, mtu int32) (*Node, error) {
	udpConn, err := newUDPConn(net, addr, mtu)
	if err != nil {
		return nil, err
	}
	
	return CreateNode(udpConn)
}

// Calls a procedure on a node using the UDP protocol
// see Node.Call
func (node *Node) CallUDP(procedure string, addrString string, args interface{}, reply interface{}, timeout int) (err error) {
	addr, err := net.ResolveUDPAddr("udp", addrString)
	if err != nil {
		return err
	}

	return node.Call(procedure, addr, args, reply, timeout)
}
