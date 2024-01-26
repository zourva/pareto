package jsonrpc2

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strconv"
)

const (
	Version = "2.0"
)

// RPCRequest represents a JSON-RPC request object.
//
//	Method: string containing the method to be invoked
//	Params: nil or a json object or an json array
//	ID: message id used to identify request and response pairs.
//	    Should be unique for every request in a batch request.
//	Version: must always be set to "2.0" for JSON-RPC version 2.0
//	See: http://www.jsonrpc.org/specification#request_object
type RPCRequest struct {
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Version string `json:"jsonrpc"`
	Params  any    `json:"params,omitempty"`
}

func (r *RPCRequest) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// GetObject returns params part of a request and convert
// it to the type instance passed-in.
func (r *RPCRequest) GetObject(toType any) error {
	msg, err := json.Marshal(r.Params)
	if err != nil {
		return NewError(ErrServerInvalidParameters)
	}

	err = json.Unmarshal(msg, toType)
	if err != nil {
		return NewError(ErrServerInvalidParameters)
	}

	return nil
}

func (r *RPCRequest) String() string {
	if r == nil {
		return "nil"
	}

	buf := fmt.Sprintf("Version: %v, ID: %v, Method: %v", r.Version, r.ID, r.Method)

	if r.Params != nil {
		buf = fmt.Sprintf("%s, Params: %v", buf, r.Params)
	}

	return buf
}

// NewRequest creates a new RPCRequest with the given message id.
func NewRequest(id int, method string, params ...any) *RPCRequest {
	request := &RPCRequest{
		ID:      id,
		Method:  method,
		Params:  parseParams(params...),
		Version: Version,
	}

	return request
}

// ParseRequest parses data to get an RPCRequest object.
func ParseRequest(data []byte) (*RPCRequest, *RPCError) {
	request := &RPCRequest{}
	err := json.Unmarshal(data, request)
	if err != nil {
		log.Warnln("rpc request params error:", err)
		return nil, NewError(ErrServerInvalidParameters)
	}

	if request.ID == 0 || request.Method == "" {
		return nil, NewError(ErrServerInvalidParameters)
	}

	return request, nil
}

// RPCResponse represents a JSON-RPC response object.
//
//	Result: holds the result of the rpc call if no error occurred, nil otherwise. can be nil even on success.
//	Error: holds an RPCError object if an error occurred. must be nil on success.
//	ID: may always be 0 for single requests. is unique for each request in a batch call (see CallBatch())
//	Version: must always be set to "2.0" for JSON-RPC version 2.0
//	See: http://www.jsonrpc.org/specification#response_object
type RPCResponse struct {
	ID      int       `json:"id"`
	Version string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

func (r *RPCResponse) String() string {
	if r == nil {
		return "nil"
	}

	buf := fmt.Sprintf("Version: %v, ID: %v", r.Version, r.ID)

	if r.Result != nil {
		buf = fmt.Sprintf("%s, Result: %v", buf, r.Result)
	}

	if r.Error != nil {
		buf = fmt.Sprintf("%s, Error: %v", buf, r.Error)
	}

	return buf
}

// GetInt converts the rpc response to an int64 and returns it.
//
// If result was not an integer an error is returned.
func (r *RPCResponse) GetInt() (int64, error) {
	val, ok := r.Result.(json.Number)
	if !ok {
		return 0, fmt.Errorf("could not parse int64 from %s", r.Result)
	}

	i, err := val.Int64()
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetFloat converts the rpc response to float64 and returns it.
//
// If result was not a float64 an error is returned.
func (r *RPCResponse) GetFloat() (float64, error) {
	val, ok := r.Result.(json.Number)
	if !ok {
		return 0, fmt.Errorf("could not parse float64 from %s", r.Result)
	}

	f, err := val.Float64()
	if err != nil {
		return 0, err
	}

	return f, nil
}

// GetBool converts the rpc response to a bool and returns it.
//
// If result was not a bool an error is returned.
func (r *RPCResponse) GetBool() (bool, error) {
	val, ok := r.Result.(bool)
	if !ok {
		return false, fmt.Errorf("could not parse bool from %s", r.Result)
	}

	return val, nil
}

// GetString converts the rpc response to a string and returns it.
//
// If result was not a string an error is returned.
func (r *RPCResponse) GetString() (string, error) {
	val, ok := r.Result.(string)
	if !ok {
		return "", fmt.Errorf("could not parse string from %s", r.Result)
	}

	return val, nil
}

// GetObject converts the rpc response to an arbitrary type.
//
// The function works as you would expect it from json.Unmarshal()
func (r *RPCResponse) GetObject(toType any) error {
	js, err := json.Marshal(r.Result)
	if err != nil {
		return NewError(ErrServerInvalidParameters)
	}

	err = json.Unmarshal(js, toType)
	if err != nil {
		return NewError(ErrServerInvalidParameters)
	}

	return nil
}

func (r *RPCResponse) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		log.Debug("jsonrpc response marshal error:", err)
		return nil, err
	}

	return b, nil
}

// NewResponse creates a response for the given request
// and payload, if any.
func NewResponse(request *RPCRequest, data any) *RPCResponse {
	response := &RPCResponse{
		ID:      request.ID,
		Result:  data,
		Version: Version,
	}

	return response
}

// NewErrorResponse creates an error response
// with the provided code and error message.
func NewErrorResponse(code int, msg string) *RPCResponse {
	if msg == "" {
		msg2, ok := ErrCodeString[code]
		if !ok {
			msg = "unknown error"
		} else {
			msg = msg2
		}
	}

	response := &RPCResponse{
		Error:   NewErrorWithMsg(code, msg),
		Version: Version,
	}

	return response
}

// ParseResponse parses data to get an RPCResponse object.
func ParseResponse(data []byte) (*RPCResponse, error) {
	response := &RPCResponse{}
	err := json.Unmarshal(data, response)
	if err != nil {
		log.Debug("rpc request params error:", err)
		return nil, NewError(ErrServerInvalidParameters)
	}

	return response, nil
}

// RPCError represents a JSON-RPC error object if an RPC error occurred.
//
//	Code: holds the error code
//	Message: holds a short error message
//	Data: holds additional error data, may be nil
//	See: http://www.jsonrpc.org/specification#error_object
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error function is provided to be used as error object.
func (e *RPCError) Error() string {
	return strconv.Itoa(e.Code) + ":" + e.Message
}

func NewError(code int) *RPCError {
	return &RPCError{
		Code:    code,
		Message: ErrCodeString[code],
	}
}

func NewErrorWithMsg(code int, msg string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: msg,
	}
}

func parseParams(params ...any) any {
	if params == nil {
		return nil
	}

	var finalParams any
	switch len(params) {
	case 0:
		return nil // returns nil when no params provided
	case 1: // assume an JSON object or a single element array
		if params[0] != nil {
			var typeOf reflect.Type

			// skip non-nil pointers
			for typeOf = reflect.TypeOf(params[0]); typeOf != nil && typeOf.Kind() == reflect.Ptr; typeOf = typeOf.Elem() {
			}

			if typeOf != nil {
				// now check if we can directly marshal the type or if it must be wrapped in an array
				switch typeOf.Kind() {
				// for these types we just do nothing, since value of p is already unwrapped from the array params
				case reflect.Struct:
					finalParams = params[0]
				case reflect.Array:
					finalParams = params[0]
				case reflect.Slice:
					finalParams = params[0]
				case reflect.Interface:
					finalParams = params[0]
				case reflect.Map:
					finalParams = params[0]
				default: // everything else must stay in an array (int, string, etc)
					finalParams = params
				}
			}
		} else {
			finalParams = params
		}
	default: // if more than one parameter was provided, treat as an array
		finalParams = params
	}

	return finalParams
}

//-32700 ---> parse error. not well formed
//-32701 ---> parse error. unsupported encoding
//-32702 ---> parse error. invalid character for encoding
//-32600 ---> server error. invalid xml-rpc. not conforming to spec.
//-32601 ---> server error. requested method not found
//-32602 ---> server error. invalid method parameters
//-32603 ---> server error. internal xml-rpc error
//-32500 ---> application error
//-32400 ---> system error
//-32300 ---> transport error

const (
	ErrParseNotWellFormed            = 32700
	ErrParseUnsupportedEncoding      = 32701
	ErrParseInvalidCharacterEncoding = 32702
	ErrServerInvalid                 = 32600 //server side(callee side) error
	ErrServerMethodNotFound          = 32601
	ErrServerInvalidParameters       = 32602
	ErrServerInternal                = 32603
	ErrServerInvalidMessageId        = 32604
	ErrApplicationError              = 32500 //application side(caller side) error
	ErrSystemError                   = 32400
	ErrTransportError                = 32300
)

var ErrCodeString = map[int]string{
	ErrParseNotWellFormed:            "parse error: message malformed",
	ErrParseUnsupportedEncoding:      "parse error: unsupported encoding",
	ErrParseInvalidCharacterEncoding: "parse error: invalid character for encoding",
	ErrServerInvalid:                 "server error: invalid rpc, not conforming to the spec",
	ErrServerMethodNotFound:          "server error: requested method not found",
	ErrServerInvalidParameters:       "server error: invalid method parameters",
	ErrServerInternal:                "server error: internal rpc error",
	ErrServerInvalidMessageId:        "server error: invalid message id",
	ErrApplicationError:              "application error",
	ErrSystemError:                   "system error",
	ErrTransportError:                "transport error",
}
