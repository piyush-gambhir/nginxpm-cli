package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	switch s {
	case "true", "1":
		*b = true
	case "false", "0", "null", "":
		*b = false
	default:
		return fmt.Errorf("cannot unmarshal %s into bool", string(data))
	}
	return nil
}

type Response struct {
	HTTPResponse *http.Response
	RequestURL   string
}

func (r *Response) JSON(v interface{}) error {
	defer r.HTTPResponse.Body.Close()

	body, err := readAllLimited(r.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if r.HTTPResponse.StatusCode >= 400 {
		msg := extractErrorMessage(body)
		return &APIError{
			StatusCode: r.HTTPResponse.StatusCode,
			Message:    msg,
			URL:        r.RequestURL,
		}
	}

	if v != nil && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

func (r *Response) Error() error {
	defer r.HTTPResponse.Body.Close()

	if r.HTTPResponse.StatusCode >= 400 {
		body, err := readAllLimited(r.HTTPResponse.Body)
		if err != nil {
			return &APIError{
				StatusCode: r.HTTPResponse.StatusCode,
				Message:    fmt.Sprintf("failed to read error body: %v", err),
				URL:        r.RequestURL,
			}
		}
		msg := extractErrorMessage(body)
		return &APIError{
			StatusCode: r.HTTPResponse.StatusCode,
			Message:    msg,
			URL:        r.RequestURL,
		}
	}
	return nil
}

func (r *Response) StatusCode() int {
	return r.HTTPResponse.StatusCode
}

func (r *Response) RawBody() ([]byte, error) {
	defer r.HTTPResponse.Body.Close()
	return readAllLimited(r.HTTPResponse.Body)
}

func extractErrorMessage(body []byte) string {
	var npmErr struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &npmErr) == nil && npmErr.Error.Message != "" {
		return npmErr.Error.Message
	}

	var msgErr struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &msgErr) == nil && msgErr.Message != "" {
		return msgErr.Message
	}

	var simpleErr struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &simpleErr) == nil && simpleErr.Error != "" {
		return simpleErr.Error
	}

	msg := string(body)
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	return msg
}
