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

// Basic wire types (message, query, reply)

import (
	"bytes"
	"github.com/zeebo/bencode"
)

type Message struct {
	Query *Query `bencode:"Q,omitempty"`
	Reply *Reply `bencode:"R,omitempty"`
}

type Query struct {
	ProcedureName string             `bencode:"N"` // Name of the procedure being executed
	ProcedureData bencode.RawMessage `bencode:"D"` // Procedure argument(s)
	MessageID     uint32             `bencode:"I"` // ID to make this transaction unique
}

func (query *Query) constructArgs(args interface{}) error {
	buf := new(bytes.Buffer)
	bencodeE := bencode.NewEncoder(buf)
	if err := bencodeE.Encode(args); err != nil {
		return err
	}

	data, err := encodeIntoBigEndian(buf)
	query.ProcedureData = data

	return err
}

type Reply struct {
	ReturnData bencode.RawMessage `bencode:"D"` // Procedure return value(s)
	MessageID  uint32             `bencode:"I"` // ID of query we are responding to
}
