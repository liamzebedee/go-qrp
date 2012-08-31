package qrp

import (
	"testing"
)

type AddService struct {}
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
	err, node := CreateNodeUDP("127.0.0.1:50000")
	if err != nil {
		print("SERVER: Can't create server node - ")
		print(err.Error())
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
	} ()
	println("SERVER: Serving")
	
	
	// Client
	err, client := CreateNodeUDP("127.0.0.1:60000")
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
	} ()
	println("CLIENT: Serving")
	
	args := AddArgs { 2, 2 }
	reply := new(AddReply)
	println("CLIENT: Calling Add on server")
	err = client.CallUDP("Add", "127.0.0.1:50000", args, reply, 3)
	go client.CallUDP("Add", "127.0.0.1:50000", args, reply, 3)
	go client.CallUDP("Add", "127.0.0.1:50000", args, reply, 3)
	go client.CallUDP("Add", "127.0.0.1:50000", args, reply, 3)
	go client.CallUDP("Add", "127.0.0.1:50000", args, reply, 3)
	
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
