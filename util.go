package qrp

import (
	"reflect"
	"unicode"
	"unicode/utf8"
	"bytes"
	"encoding/binary"
	"fmt"
)

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Decode data into Big Endian encoding
func decodeIntoBigEndian(data *bytes.Buffer) ([]byte, error) {
	buf_bigEndian := make([]byte, data.Len())
	err := binary.Read(data, binary.BigEndian, buf_bigEndian) // Read data into buf_bigEndian
	fmt.Print("")
	return buf_bigEndian, err
}

// Encode data into Big Endian encoding
func encodeIntoBigEndian(data *bytes.Buffer) ([]byte, error) {
	buf_bigEndian := new(bytes.Buffer) // Buffer to hold BigEndian payload
	err := binary.Write(buf_bigEndian, binary.BigEndian, data.Bytes()) // Write data into buf_bigEndian
	if err != nil {
		return nil, err
	}
	
	return buf_bigEndian.Bytes(), nil
}