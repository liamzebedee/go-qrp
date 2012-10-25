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

func Test_UDP(t *testing.T) {
	println("=== Running UDP Test")
	
	// Server
	server, err := CreateNodeUDP("udp", "127.0.0.1:50000", 512)
	if err != nil {
		println("SERVER: Can't create server node -", err.Error())
		t.FailNow()
	}
	println("SERVER: Server created")

	add := new(AddService)
	server.Register(add)
	println("SERVER: Add service registered")
	
	go func() {
		err := server.ListenAndServe()
		
		if err != nil {
			println("SERVER:", "Error serving -", err.Error())
			t.FailNow()
		}
	} ()
	
	println("SERVER: Serving")
	//defer server.Stop()

	// Client
	client, err := CreateNodeUDP("udp", "127.0.0.1:60000", 512)
	if err != nil {
		println("CLIENT: Can't create client node - ", err.Error())
		t.FailNow()
	}

	println("CLIENT: Client node created")
	go func() {
		err := client.ListenAndServe()
		
		if err != nil {
			println("CLIENT:", "Error serving -", err.Error())
			t.FailNow()
		}
	}()
	
	println("CLIENT: Serving")
	//defer client.Stop()

	args := AddArgs{2, 2}
	reply := new(AddReply)
	println("CLIENT: Calling Add on server")
	err = client.CallUDP("Add", "127.0.0.1:50000", args, reply, 1)
	
	if err != nil {
		println("CLIENT: Client call error - ", err.Error())
		t.FailNow()
	}
	if reply != nil {
		print("CLIENT: 2+2=")
		println(reply.Result)
	}
	
	println("UDP test succeeded!")
	println("")
}
/*
func Test_TCP(t *testing.T) {
	println("=== Running TCP Test")

	// Server
	server, err := CreateNodeTCP("tcp", "127.0.0.1:50000", "127.0.0.1:50001")
	if err != nil {
		print("SERVER: Can't create server node -", err.Error())
		t.FailNow()
	}
	println("SERVER: Server created")

	add := new(AddService)
	server.Register(add)
	println("SERVER: Add service registered")

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			println("SERVER:", "Error serving -", err.Error())
			t.FailNow()
		}
	}()
	println("SERVER: Serving")
	defer server.Stop()

	// Client
	client, err := CreateNodeTCP("tcp", "127.0.0.1:60000", "127.0.0.1:60001")
	if err != nil {
		print("CLIENT: Can't create client node - ", err.Error())
		t.FailNow()
	}

	println("CLIENT: Client node created")
	go func() {
		err := client.ListenAndServe()
		if err != nil {
			println("CLIENT:", "Error serving -", err.Error())
			t.FailNow()
		}
	}()
	println("CLIENT: Serving")
	defer client.Stop()

	args := AddArgs{2, 2}
	reply := new(AddReply)
	println("CLIENT: Calling Add on server")
	err = client.CallTCP("Add", "127.0.0.1:50000", args, reply, 4)

	if err != nil {
		print("CLIENT: Client call error - ")
		println(err.Error())
		t.FailNow()
	}
	if reply != nil {
		print("CLIENT: 2+2=")
		println(reply.Result)
	}

	println("TCP test succeeded!")
	println("")
}*/