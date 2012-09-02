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

func (conn *UDPConn) ReadNextPacket() (buffer []byte, read int, addr net.Addr, err error) {
	// Buffer size is 512 because it's the largest size without possible fragmentation
	buffer = make([]byte, conn.mtu)

	// Read a packet into the buffer
	read, addr, err = conn.ReadFrom(buffer)

	return buffer, read, addr, err
}
