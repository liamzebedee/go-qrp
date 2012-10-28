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

// Utility functions

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"unicode"
	"unicode/utf8"
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
	if err != nil {
		return nil, err
	}

	return buf_bigEndian, nil
}

// Encode data into Big Endian encoding
func encodeIntoBigEndian(data *bytes.Buffer) ([]byte, error) {
	buf_bigEndian := new(bytes.Buffer)                                 // Buffer to hold BigEndian payload
	err := binary.Write(buf_bigEndian, binary.BigEndian, data.Bytes()) // Write data into buf_bigEndian
	if err != nil {
		return nil, err
	}

	return buf_bigEndian.Bytes(), nil
}
