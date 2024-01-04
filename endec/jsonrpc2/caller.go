package jsonrpc2

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"time"
)

// Bearer defines underlying calling process manager
// of JSON-RPC implementation.
type Bearer interface {
	Call(channel string, data []byte, to time.Duration) ([]byte, error)
}

// Client defines a general JSON-RPC method caller.
type Client struct {
	bearer Bearer
}

func (i *Client) getId() int {
	return rand.Intn(math.MaxUint32)
}

// Invoke invokes a method of the given name with a timeout specified, and parameters provided.
//
//		e.g.: Invoke("test.string", 1*time.Second, "hello", "json-rpc")
//	       Invoke("test.struct", 2*time.Second, someStruct)
func (i *Client) Invoke(channel, method string, timeout time.Duration, params ...any) (*RPCResponse, error) {
	req := NewRequest(i.getId(), method, params...)

	reqBuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	rspBuf, err := i.bearer.Call(channel, reqBuf, timeout)
	if err != nil {
		return nil, err
	}

	rsp, err := ParseResponse(rspBuf)
	if err != nil {
		return nil, err
	}

	// validate message id
	if rsp.ID != req.ID {
		return nil, errors.New(ErrCodeString[ErrServerInvalidMessageId])
	}

	if rsp.Error != nil && rsp.Error.Error() != "" {
		return nil, rsp.Error
	}

	return rsp, nil
}

// NewClient creates an RPC method invoker
// using the given bearer.
func NewClient(bearer Bearer) *Client {
	return &Client{
		bearer: bearer,
	}
}
