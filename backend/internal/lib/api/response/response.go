// Package response provides standardized JSON response structures and helpers
// for HTTP handlers in the WB backend API.
package response

// Response is a generic JSON response structure used in API endpoints.
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	statusOK = "OK" 
	statusError = "Error"
)

// OK returns a successful response with status "OK" and no error.
func OK() Response {
	return Response{
		Status: statusOK,
	}
}

// Error returns a failed response with status "Error" and provided message.
func Error(msg string) Response {
	return Response{
		Status: statusError,
		Error: msg,
	}
}