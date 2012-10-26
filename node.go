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

// Handles packets and overall operations of the node
// TODO
// - Make logging better, more detailed (IP+port of node) and uniform
// - TCP

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/zeebo/bencode"
	"net"
	"reflect"
	"sync"
	"time"
)

type responseChannel chan bencode.RawMessage

// A procedure that can be invoked by other machines
type procedure struct {
	Method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	Receiver  reflect.Value
}

type call struct {
	MessageID uint32
	Addr      string // Necessary because maps cannot compare interfaces (net.Addr)
}

// Our local node
type Node struct {
	Connection
	procedures map[string]*procedure     // Registered procedures on the node
	pending    map[call]*responseChannel // A map of calls to queries pending responses
	messageID  uint32                    // Current messageID
	
	pendingMutex    sync.Mutex // to protect pending, messageID
	sendingMutex    sync.Mutex
	proceduresMutex sync.Mutex
}

// Protocol specific implementation (e.g. udp.go)
type Connection interface {
	// Listens for queries and replies, serving procedures registered by Register. 
	// Returns an error if there was a failure serving or we are already serving
	ListenAndServe() (err error) 

	// WriteTo writes a packet with payload b to addr.
	WriteTo(b []byte, addr net.Addr) (n int, err error)

	// Stops the server. Returns an error if the server isn't running.
	Stop() error
}

// Creates a node that performs IO on connection
func CreateNode() (Node) {
	node := Node{}

	// Maps
	node.procedures = make(map[string]*procedure)
	node.pending = make(map[call]*responseChannel)

	// Mutices
	node.sendingMutex = *new(sync.Mutex)
	node.pendingMutex = *new(sync.Mutex)
	node.proceduresMutex = *new(sync.Mutex)
	
	// Initialize messageID to pseudorandom value
	node.messageID = uint32(time.Now().Nanosecond())

	//log.SetPrefix(strings.Join([]string{ "qrp - [", addr.String(), "]:" }, ""))
	
	return node
}

type packet struct {
	buffer []byte
	read int
	addr net.Addr
	err error
}

// Processes received packets
func (node *Node) processPacket(data []byte, read int, addr net.Addr) error {
	data_bigEndian, err := decodeIntoBigEndian(bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// Unmarshal BigEndian BEncode into struct
	bencodeDecoder := bencode.NewDecoder(bytes.NewBuffer(data_bigEndian))
	var message Message
	if err := bencodeDecoder.Decode(&message); err != nil {
		return err
	}

	// Further processing
	return node.processMessage(&message, addr)
}

// Processes raw messages
func (node *Node) processMessage(message *Message, addr net.Addr) error {
	if query := message.Query; query != nil {
		// Query
		return node.processQuery(query, addr)
	} else if reply := message.Reply; reply != nil {
		// Reply
		return node.processReply(reply, addr)
	} else {
		return new(InvalidMessageError)
	}
	return nil
}

// Processes received queries
func (node *Node) processQuery(query *Query, addr net.Addr) error {
	procedureName := query.ProcedureName
	if procedure := node.procedures[procedureName]; procedure != nil {
		// Procedure exists (it's being served)
		method := procedure.Method
		function := method.Func

		// Initialize value
		argValue, replyValue := reflect.New(procedure.ArgType.Elem()), reflect.New(procedure.ReplyType.Elem())

		// Set value of arg
		argsReader := bytes.NewReader(query.ProcedureData)
		argsDecoder := bencode.NewDecoder(argsReader)
		err := argsDecoder.Decode(argValue.Interface())
		if err != nil {
			fmt.Errorf("qrp:", "Error decoding procedure data into value: %s\n", err.Error())
			return err
		}

		// Invoke the function
		function.Call([]reflect.Value{procedure.Receiver, argValue, replyValue})

		// Create reply
		reply := Reply{MessageID: query.MessageID}
		argsBuf := new(bytes.Buffer)
		argsEncoder := bencode.NewEncoder(argsBuf)
		argsEncoder.Encode(replyValue.Interface())
		reply.ReturnData, err = encodeIntoBigEndian(argsBuf)

		if err != nil {
			fmt.Errorf("qrp:", "Error encoding reply return data: %s\n", err.Error())
			return err
		}

		// Create message
		message := Message{Reply: &reply}

		// Encode message
		messageBuf := new(bytes.Buffer)
		messageEncoder := bencode.NewEncoder(messageBuf)
		err = messageEncoder.Encode(message)
		if err != nil {
			fmt.Errorf("qrp:", "Error encoding reply message into BEncode: %s\n", err.Error())
			return err
		}

		message_bigEndian, err := encodeIntoBigEndian(messageBuf)
		if err != nil {
			fmt.Errorf("qrp:", "Error encoding reply message into BigEndian: %s\n", err.Error())
			return err
		}

		// Send to host
		node.sendingMutex.Lock()
		// TODO: Make safe from server closedown
		var _ = message_bigEndian
		if _, err := node.WriteTo(message_bigEndian, addr); err != nil {
			return err
		}
		node.sendingMutex.Unlock()

		return nil
	} else {
		return &BadProcedureError{procedureName}
	}
	return nil
}

// Processes received replies
func (node *Node) processReply(reply *Reply, addr net.Addr) error {
	// Construct call
	chanCall := call{MessageID: reply.MessageID, Addr: addr.String()}

	// Get associated channel
	node.pendingMutex.Lock()
	responseChan := node.pending[chanCall] // Get response channel
	node.pendingMutex.Unlock()

	if responseChan == nil {
		return &InvalidMessageMappingError{reply.MessageID}
	}

	// Send return data
	(*responseChan) <- reply.ReturnData

	return nil
}

// Returns the next available call slot for an IP
// The ID space is unique to 2 communicating nodes
// This is managed by maintaining a map of IDs to Calls. A call contains an IP+ID combination for the call. Only IDs that come from the same IP can be mapped to calls 
func (node *Node) nextCall(addr net.Addr) (nextCall call) {
	node.pendingMutex.Lock()
	defer node.pendingMutex.Unlock()

	// True when we have found an ID which is free
	// In ~99.999999999% of circumstances, this will run once
	callCreateDone := false

	for !callCreateDone {
		// Go doesn't panic after integer overflow, so this is OKAY!
		node.messageID++
		nextCall = call{MessageID: node.messageID, Addr: addr.String()}

		// If there isn't already a pending call with the same IP+ID combination
		if node.pending[nextCall] == nil {
			callCreateDone = true
		}
	}

	return nextCall
}

// Tries to call 'procedure' on remote node, with supplied 'args' and allocated return values 'reply'. 
// 'timeout' can be used to specify a maximum time to wait for a reply (in seconds). If timeout is 0, we wait forever. 
// The reliability of this completing successfully is dependent on the network protocol (UDP is unreliable)
// Returns an error if there is a timeout
func (node *Node) Call(procedure string, addr net.Addr, args interface{}, reply interface{}, timeout int) (err error) {
	// Get our call, which contains the message ID
	call := node.nextCall(addr)

	// Create Query
	query := Query{ProcedureName: procedure, MessageID: call.MessageID}
	query.constructArgs(args)

	// Create Message
	message := Message{Query: &query}

	// Encode it into BEncode
	buf := new(bytes.Buffer)
	bencodeE := bencode.NewEncoder(buf)
	if err := bencodeE.Encode(message); err != nil {
		fmt.Errorf("qrp:", "Error encoding query message into BEncode: %s\n", err.Error())
		return err
	}

	buf_bigEndian, err := encodeIntoBigEndian(buf)
	if err != nil {
		fmt.Errorf("qrp:", "Error encoding query message into BigEndian: %s\n", err.Error())
		return err
	}

	// Create channel for receiving response
	responseChan := make(responseChannel)

	// Allocate channel
	node.pendingMutex.Lock()
	node.pending[call] = &responseChan
	node.pendingMutex.Unlock()

	// Delete channel after exit
	defer func() {
		node.pendingMutex.Lock()
		delete(node.pending, call)
		node.pendingMutex.Unlock()
	}()

	// Send to host
	node.sendingMutex.Lock()
	// TODO: Make safe from server closedown
	if _, err = node.WriteTo(buf_bigEndian, addr); err != nil {
		return err
	}
	node.sendingMutex.Unlock()
	
	// If timeout isn't 0, initate the timeout function concurrently
	timeoutChan := make(chan bool, 1)
	if timeout > 0 {
		go func() {
			// Timeout function
			time.Sleep(time.Duration(timeout) * time.Second)
			timeoutChan <- true
		}()
	}

	// Wait for response on channel
	select {
		case replydata := <-responseChan:
			// We received a reply
			// Decode args
			argsReader := bytes.NewReader(replydata)
			argsDecoder := bencode.NewDecoder(argsReader)
			err := argsDecoder.Decode(reply)
			if err != nil {
				fmt.Errorf("qrp:", "Error decoding reply return data into value: %s\n", err.Error())
				return err
			}
		case <-timeoutChan:
			// We timed out
			return new(TimeoutError)
	}

	return nil
}

// Registers method as a procedure, which must satisfy the following conditions:
//	- exported
//  - has a receiver
//	- two arguments, both pointers to exported structs
//	- one return value, of type error
// It returns an error if the method does not satisfy these conditions
func (node *Node) Register(receiver interface{}) error {
	return node.register(receiver)
}

// Registers a method that
func (node *Node) register(receiver interface{}) error {
	// Lock mutex, prevents state corruption
	node.proceduresMutex.Lock()

	// Create service map if not made already
	if node.procedures == nil {
		node.procedures = make(map[string]*procedure)
	}

	// Declarations
	argIndex, replyIndex := 1, 2
	// Method needs two/three ins: receiver, *args, *reply.
	maxIns := 3
	receiverType := reflect.TypeOf(receiver)

	// Install the methods
	for m := 0; m < receiverType.NumMethod(); m++ {
		method := receiverType.Method(m)
		procedure := new(procedure)
		methodType := method.Type
		methodName := method.Name

		// TODO: Maybe replace the error handling with something more friendly?
		var errorBuf bytes.Buffer
		throwError := func() error {
			fmt.Errorf(errorBuf.String())
			return errors.New(errorBuf.String())
		}

		if methodType.NumIn() != maxIns {
			fmt.Fprintln(&errorBuf, "method", methodName, "has wrong number of ins:", methodType.NumIn())
			throwError()
		}

		// First arg need not be a pointer.
		argType := methodType.In(argIndex)
		if !isExportedOrBuiltinType(argType) {
			fmt.Fprintln(&errorBuf, methodName, "argument type not exported:", argType)
			throwError()
		}

		// Second arg must be a pointer.
		replyType := methodType.In(replyIndex)
		if replyType.Kind() != reflect.Ptr {
			fmt.Fprintln(&errorBuf, "method", methodName, "reply type not a pointer:", replyType)
			throwError()
		}

		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			fmt.Fprintln(&errorBuf, "method", methodName, "reply type not exported:", replyType)
			throwError()
		}

		// Register method
		procedure.Method = method
		procedure.ArgType = argType
		procedure.ReplyType = replyType
		procedure.Receiver = reflect.ValueOf(receiver)
		node.procedures[methodName] = procedure
	}
	node.proceduresMutex.Unlock()
	return nil
}
