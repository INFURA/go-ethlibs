package jsonrpc

import (
	"encoding/json"
)

// Currently NotificationParams are always a JSON object, but this may change, in which
// case the code around NotificationParams will need to be updated.
type NotificationParams = json.RawMessage

type Notification struct {
	JSONRPC string             `json:"jsonrpc"`
	Method  string             `json:"method"`
	Params  NotificationParams `json:"params"`
}

// MarshalJSON implements json.Marshaler and adds the "jsonrpc":"2.0"
// property.
func (n Notification) MarshalJSON() ([]byte, error) {
	n2 := struct {
		JSONRPC string             `json:"jsonrpc"`
		Method  string             `json:"method"`
		Params  NotificationParams `json:"params"`
	}{
		JSONRPC: "2.0",
		Method:  n.Method,
		Params:  n.Params,
	}
	return json.Marshal(n2)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Notification) UnmarshalJSON(data []byte) error {
	type tmpType Notification

	if err := json.Unmarshal(data, (*tmpType)(n)); err != nil {
		return err
	}
	return nil
}

// UnmarshalParamsInto will decode NotificationParams into the passed in value, which
// must be a pointer receiver.  The type of the passed in value is used to Unmarshal the data.
// UnmarshalInto will fail if the parameters cannot be converted to the passed-in types.
//
// Example:
//
//   var newHead eth.NewHeads
//   err := notification.UnmarshalParamsInto(&newHead)
//
// IMPORTANT: While Go will compile with non-pointer receivers, the Unmarshal attempt will
// *always* fail with an error.
func (n *Notification) UnmarshalParamsInto(receiver interface{}) error {
	return json.Unmarshal(n.Params, receiver)
}
