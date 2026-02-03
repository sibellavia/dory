package plugin

// Request is a single JSON-RPC-like request line sent to a plugin process.
type Request struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Response is a single JSON-RPC-like response line from a plugin process.
type Response struct {
	ID     string                 `json:"id"`
	Result map[string]interface{} `json:"result,omitempty"`
	Error  *ResponseError         `json:"error,omitempty"`
}

// ResponseError is a structured plugin response error.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
