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

/*
Package qrp provides access to the exported methods of an object across a packet-based connection.
QRP is designed to work with multiple packet-based protocols such as UDP and SCTP. 

Firstly you must create a node using any of the network specific CreateNode? methods:
 CreateNode
 CreateNodeUDP

e.g.
 	node, err := CreateNodeUDP("udp", ":50060", 512)

Nodes serve procedures. Procedures are installed using the Register function, which takes an object and
registers the exported methods of that object.

 	type AddService struct {}
 	
 	type AddArgs struct {
		A, B int32
	}
	type AddReply struct {
		Result int32
	}
	
	func (s *AddService) Add(args *AddArgs, reply *AddReply) {
		reply.Result = args.A + args.B
	}

	node.Register(AddService{})
	

For more infomation on how to use QRP, see rpc_test.go. Most of this library follows the same standards in go/net/rpc.

For more information on the design of the protocol, see doc.go.
*/
package qrp

/*
QRP: A simple packet-based RPC protocol

QRP is a simple efficient protocol for short and simple two-message communications between nodes

The protocol consists of nodes communicating via BEncoded messages. A message is either
a query or a reply. 

Each message is structured so a top-level key identifies the message type and further data is
mapped to this. This key is either 'Q' for query or 'R' for reply. 

A message with a key 'Q' is a query. A query contains 3 keys. 'N' maps to the name of the procedure
being called. 'D' maps to a list of arguments. 'I' maps to the uint32 transaction ID, used to make 
the transaction unique.

The other type of message is a reply, which has a key 'R'. A reply contains 2 keys.
The first key is 'R', which maps to the return data. The other key is 'I', which is
a uint32 used to identify the query the reply is responding to. 

On stream-based connections, QRP delimits messages by the NUL byte (0x00)

Why QRP? Why not JSON-RPC?

QRP is designed to be really really minimal. JSON-RPC uses the full names of properties (method 
instead of m etc.). JSON itself is also quite a bulky encoding scheme in comparison to BEncode. 
Another difference between JSON-RPC and QRP is the absence of an error property. Errors aren't 
always useful for a node, so why include them?

QRP is also very modular in comparison to go/net/rpc. node.go contains all the functions related to
operating a node, including processing received packets 
*/
