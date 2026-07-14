package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Bool is a bool that tolerates the different encodings used by the
// Nginx Proxy Manager API. NPM releases before v2.12.0 return database
// boolean columns as JSON numbers (0/1); newer releases return real
// booleans. Bool accepts true/false, 0/1, "0"/"1"/"true"/"false", and null.
type Bool bool

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bool) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	// Unwrap string-encoded values like "1" or "true".
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

// Response wraps an http.Response with convenience methods.
type Response struct {
	HTTPResponse *http.Response
	RequestURL   string
}

// JSON reads the response body and unmarshals it into v.
// It also checks for error status codes.
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

// Error checks if the response indicates an error and returns an APIError.
// It always closes the response body to prevent leaks.
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

// StatusCode returns the HTTP status code.
func (r *Response) StatusCode() int {
	return r.HTTPResponse.StatusCode
}

// RawBody reads and returns the raw response body.
func (r *Response) RawBody() ([]byte, error) {
	defer r.HTTPResponse.Body.Close()
	return readAllLimited(r.HTTPResponse.Body)
}

// extractErrorMessage attempts to extract a meaningful error message from
// an NPM error response body.
func extractErrorMessage(body []byte) string {
	// NPM errors: {"error":{"code":400,"message":"..."}}
	var npmErr struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &npmErr) == nil && npmErr.Error.Message != "" {
		return npmErr.Error.Message
	}

	// Try simple message field: {"message":"..."}
	var msgErr struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &msgErr) == nil && msgErr.Message != "" {
		return msgErr.Message
	}

	// Try simple string error: {"error":"..."}
	var simpleErr struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &simpleErr) == nil && simpleErr.Error != "" {
		return simpleErr.Error
	}

	// Fall back to raw body.
	msg := string(body)
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	return msg
}
