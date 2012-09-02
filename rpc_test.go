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

// Test to show the capabilities and use of the qrp package

import (
	"testing"
)

type AddService struct{}

func (s *AddService) Add(args *AddArgs, reply *AddReply) {
	reply.Result = args.A + args.B
}

type AddArgs struct {
	A, B int32
}
type AddReply struct {
	Result int32
}

func Test_Single(*testing.T) {
	// Server
	err, node := CreateNodeUDP("udp", "127.0.0.1:50000", 512)
	if err != nil {
		print("SERVER: Can't create server node -", err.Error())
		return
	}
	println("SERVER: Server created")

	add := new(AddService)
	node.Register(add)
	println("SERVER: Add service registered")

	go func() {
		err := node.ListenAndServe()
		if err != nil {
			println("SERVER:", "Error serving -", err.Error())
			return
		}
	}()
	println("SERVER: Serving")

	// Client
	err, client := CreateNodeUDP("udp", "127.0.0.1:60000", 512)
	if err != nil {
		print("CLIENT: Can't create client node - ", err.Error())
		return
	}

	println("CLIENT: Client node created")
	go func() {
		err := client.ListenAndServe()
		if err != nil {
			println("CLIENT:", "Error serving -", err.Error())
			return
		}
	}()
	println("CLIENT: Serving")

	args := AddArgs{2, 2}
	reply := new(AddReply)
	println("CLIENT: Calling Add on server")
	err = client.CallUDP("Add", "127.0.0.1:50000", args, reply, 3)

	if err != nil {
		print("CLIENT: Client call error - ")
		println(err.Error())
		return
	}
	if reply != nil {
		print("CLIENT: 2+2=")
		println(reply.Result)
	}

	println("Finished tests...")
}
