package qrp

import (
	"testing"
)

type AddService struct {}
func (s *AddService) Add(args *AddArgs, reply *AddReply) {
	reply.result = args.a + args.b
}


type AddArgs struct {
	a, b int
}
type AddReply struct {
	result int
}

func Test_Single(*testing.T) {
	err, node := CreateNodeUDP("127.0.0.1:50000", 1024)
	if err != nil {
		print("Can't create server node")
		print(err.Error())
		return
	}
	println("Server created")
	add := new(AddService)
	node.Register(add)
	println("Add service registered")
	go func() {
		err = node.ListenAndServe()
		if err != nil {
			print(err.Error())
			return
		}
	} ()
	println("Serving")
	
	err, client := CreateNodeUDP("127.0.0.1:60000", 1024)
	if err != nil {
		print("Can't create client node")
		print(err.Error())
		return
	}
	println("Client node created")
	
	args := AddArgs { 2, 2 }
	reply := new(AddReply)
	err = client.Call("Add", "127.0.0.1:50000", args, reply, 3)
	println("Calling Add on server")
	if err != nil {
		println("Client call error")
		println(err.Error())
		return
	}
	if reply != nil {
		print("2+2=")
		println(reply.result)
	}
	
	println("Done.")
}
