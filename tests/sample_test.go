package tests

import (
	"testing"
	"github.com/liamzebedee/bencode/src/bencode" // BEncode
	"bytes"
	"fmt"
)

type Wrap struct {
	X1A *X1 `bencode:"XXXXX,omitempty"`
	X2A *X2 `bencode:"XXAS,omitempty"`
}

type X1 struct {
	Args [1]interface{}
}

type X2 struct {
	Args [1]interface{}
	AA int32
}

func Test_Single(t *testing.T) {
	print("LZ V\n")
	x1 := X1 {}
	wrap := Wrap { X1A: &x1 }
	wrap.X1A.Args[0] = 123
	buf := new(bytes.Buffer)
	bencodeE := bencode.NewEncoder(buf)
	if err := bencodeE.Encode(wrap); err != nil {
		println(err.Error())
	} else {
		fmt.Printf("Worked? %s\n", buf.Bytes())
	}
}
