package qrp

import (

)

/*

QRP: A simple connectionless RPC protocol

QRP is a simple efficient protocol for simple two-message communications between nodes

The protocol consists of nodes communicating via sending BEncoded messages. A message is either
a query or a reply. 

Every message has an 'M' key, whose value indicates whether it is a query, 'Q', or a reply, 'R'.

Both types of messages also have an 'I' key, which is the transaction ID, used for mapping 
replies to queries. For a query, this ID can be any 32bit integer. For a reply, this ID is
what it is in response to. 

= Query =
A query has 2 key/value pairs:
 'Q':string     -> Procedure Name
 'D':dictionary -> Procedure Data (parameters)

e.g.
::JSON
 {
 	'M': 'Q',
 	'Q': 'Add',
 	'I': 21321,
 	'D': {
 		'a': 1,
 		'b': 2
 	}
 }

= Reply =
A reply has a single key/value pair:
 'D':dictionary -> Procedure Return Data

e.g.
::JSON
{
	'M': 'R',
	'I': '21321',
	'D': {
		'R': 3
	}
}

*/