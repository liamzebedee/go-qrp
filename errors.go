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

// Declarations of general errors

import (
	"fmt"
	"reflect"
	"strings"
)

// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type TimeoutError struct{}

func (err *TimeoutError) Error() string {
	return "The query waiting for a response timed out"
}

type InvalidMessageError struct{}

func (err *InvalidMessageError) Error() string {
	return "The message is invalid"
}

type BadProcedureError struct {
	procedure string
}

func (err *BadProcedureError) Error() string {
	return strings.Join([]string{"No such procedure served: ", err.procedure}, "")
}

type InvalidMessageMappingError struct {
	messageID uint32
}

func (err *InvalidMessageMappingError) Error() string {
	return fmt.Sprintf("No query mapped to message ID %d", err.messageID)
}
