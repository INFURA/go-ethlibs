package jsonrpc

import (
	"encoding/json"
)

func Unmarshal(data []byte) (interface{}, error) {
	type Unknown struct {
		Method *string          `json:"method,omitempty"`
		Params *json.RawMessage `json:"params,omitempty"`
		Result *json.RawMessage `json:"result,omitempty"`
		Error  *json.RawMessage `json:"error,omitempty"`
		ID     *json.RawMessage `json:"id"`
	}

	u := Unknown{}
	err := json.Unmarshal(data, &u)
	if err != nil {
		return nil, err
	}

	if u.Method != nil {
		// it's either a request or notification
		if u.ID != nil {
			request := Request{}
			err = json.Unmarshal(data, &request)
			if err != nil {
				return nil, err
			}
			return &request, nil
		} else {
			notif := Notification{}
			err = json.Unmarshal(data, &notif)
			if err != nil {
				return nil, err
			}
			return &notif, nil
		}
	} else {
		// it's a response
		response := RawResponse{}
		err = json.Unmarshal(data, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}
}
