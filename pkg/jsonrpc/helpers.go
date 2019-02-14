package jsonrpc

import (
	"encoding/json"
	"net/http"
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
		response := Response{}
		err = json.Unmarshal(data, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}
}

// Populate the jsonrpc.Request from a Get request and return a byte[] representation. Note that Method will need to be filled in
// outside of this function, as it could be delivered in any number of ways.
func (r *Request) FromHttpGetRequest(req *http.Request) error {
	r.JSONRPC = "2.0"
	queryParams := req.URL.Query()
	paramsQuery := queryParams.Get("params")
	if paramsQuery == "" {
		r.Params = nil
	} else {
		params := Params{}
		err := json.Unmarshal([]byte(paramsQuery), &params)
		if err != nil {
			return err
		}
		r.Params = params
	}
	return nil
}

func (r *Request) FromHttpPostRequest(req *http.Request) ([]byte, error) {
	return nil, nil
}
