package qrp

import (
	"reflect"
	"unicode"
	"unicode/utf8"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
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

type ConnBuffer bytes.Buffer

// ReadFromConn reads data from r until EOF and appends it to the buffer.
// The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read
// is also returned.
// If the buffer becomes too large, ReadFrom will panic with
// ErrTooLarge.
func (b *ConnBuffer) ReadFromConn(r net.PacketConn) (n int64, err error) {
	b.lastRead = opInvalid
	// If buffer is empty, reset to recover space.
	if b.off >= len(b.buf) {
		b.Truncate(0)
	}
	for {
		if free := cap(b.buf) - len(b.buf); free < MinRead {
			// not enough space at end
			newBuf := b.buf
			if b.off+free < MinRead {
				// not enough space using beginning of buffer;
				// double buffer capacity
				newBuf = makeSlice(2*cap(b.buf) + MinRead)
			}
			copy(newBuf, b.buf[b.off:])
			b.buf = newBuf[:len(b.buf)-b.off]
			b.off = 0
		}
		m, e := r.Read(b.buf[len(b.buf):cap(b.buf)])
		b.buf = b.buf[0 : len(b.buf)+m]
		n += int64(m)
		if e == io.EOF {
			break
		}
		if e != nil {
			return n, e
		}
	}
	return n, nil // err is EOF, so return nil explicitly
}