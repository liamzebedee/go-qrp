package qrp

import (
	"reflect"
	"strings"
)

type Message struct {
	Query *Query `bencode:"Q,omitempty"`
	Reply *Reply `bencode:"R,omitempty"`
}

type Query struct {
	ProcedureName    string       `bencode:"N"` // Name of the procedure being executed
	ProcedureData    [1]interface{}  `bencode:"D"` // Procedure argument(s)
	MessageID        uint32       `bencode:"I"` // ID to make this communication unique
}

type Reply struct {
	ReturnData [1]interface{} `bencode:"D"` // Procedure return value(s)
	MessageID  uint32                  `bencode:"I"`  // ID of query we are responding to
}

// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type TimeoutError struct {}
func (err *TimeoutError) Error() string {
	return "The query waiting for a response timed out"
}

type InvalidMessageError struct {}
func (err *InvalidMessageError) Error() string {
	return "The message is invalid"
}

type BadProcedureError struct {
	procedure string
}
func (err *BadProcedureError) Error() string {
	return strings.Join([]string{ "No such procedure served: ", err.procedure }, "")
}