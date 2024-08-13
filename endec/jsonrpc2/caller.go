package jsonrpc2

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"time"
)

// Invoker defines underlying calling process manager
// of JSON-RPC implementation.
type Invoker interface {
	Call(channel string, data []byte, to time.Duration) ([]byte, error)
}

// Client defines a general JSON-RPC method caller.
type Client struct {
	invoker Invoker
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

	rspBuf, err := i.invoker.Call(channel, reqBuf, timeout)
	if err != nil {
		return nil, err
	}

	rsp, err := ParseResponse(rspBuf)
	if err != nil {
		return nil, err
	}

	if rsp.Error != nil && rsp.Error.Error() != "" {
		return nil, rsp.Error
	}

	// validate message id
	if rsp.ID != req.ID {
		return nil, errors.New(ErrCodeString[ErrServerInvalidMessageId])
	}

	return rsp, nil
}

// NewClient creates an RPC method invoker
// using the given invoker.
func NewClient(invoker Invoker) *Client {
	if invoker == nil {
		log.Fatalln("invoker must not be nil")
	}

	return &Client{
		invoker: invoker,
	}
}
