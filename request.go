package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	LimitedStringMaxLength = 12
)

type LimitedString string

func (ls LimitedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(ls)
}

func (ls *LimitedString) UnmarshalJSON(data []byte) error {
	var s string
	var err error
	if len(data) > LimitedStringMaxLength {
		return fmt.Errorf("LimitedStringMaxLength : %v, have : %v", LimitedStringMaxLength, len(data))
	}
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ls = LimitedString(s)
	return nil
}

// Surface struct for unmarshal http.Request.Body
type PreRequest struct {
	ID     LimitedString   `json:"id,omitempty"`
	Method LimitedString   `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

func (p PreRequest) Request() *Request {
	return &Request{
		ID:     string(p.ID),
		Method: string(p.Method),
		Params: p.Params,
	}
}

// Data actually use
type Request struct {
	ID     string
	Method string
	Params json.RawMessage
}

func parseRawRequest(req *http.Request) ([]byte, error) {
	b, _ := ioutil.ReadAll(req.Body)
	var validJSON json.RawMessage
	err := json.Unmarshal(b, &validJSON)
	if err != nil {
		return nil, err
	}
	return validJSON, nil
}

func newRequest(raw json.RawMessage) (*Request, error) {
	var pre PreRequest
	// Check LimitedString
	err := json.Unmarshal(raw, &pre)
	if err != nil {
		return nil, err
	}
	return pre.Request(), nil
}

func main() {
	dataset := []string{
		// &{dbadoy something [123 34 107 101 121 34 58 34 115 111 109 101 34 125]}
		`{"id": "dbadoy","method":"something","params":{"key":"some"}}`,
		// LimitedStringMaxLength : 12, have : 13
		`{"id": "dbadoyabcde","method":"something","params":{"key":"some"}}`,
		// LimitedStringMaxLength : 12, have : 15
		`{"id": "dbadoy","method":"1234567890123","params":{"key":"some"}}`,
	}

	for _, data := range dataset {
		b := bytes.NewBuffer([]byte(data))
		req, err := http.NewRequest("POST", "http://127.0.0.1:8080", b)
		if err != nil {
			fmt.Println(err)
			return
		}
		// request.Body to []byte
		// Validation that is correct JSON format
		raw, err := parseRawRequest(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Check each content size while Unmarhsal
		a, err := newRequest(raw)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(a)
		}
	}
}
