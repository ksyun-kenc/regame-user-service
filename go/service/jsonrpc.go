package service

import (
	"encoding/json"
	"net/http"
)

const (
	rpcCodeParseErrorInvalidCharacter    = -32702
	rpcCodeParseErrorUnsupportedEncoding = -32701
	rpcCodeParseErrorNotWellFormed       = -32700

	rpcCodeInternalError  = -32603
	rpcCodeInvalidParams  = -32602
	rpcCodeMethodNotFound = -32601
	rpcCodeInvalidRequest = -32600

	rpcCodeApplicationError = -32500

	rpcCodeSystemError = -32400

	rpcCodeTransportError = -32300

	rpcCodeAuthenticationFailed = -101
	rpcCodeSessionNotFound      = -201
	rpcCodeSessionExpired       = -202
)

var rpcCodeText = map[int]string{
	rpcCodeParseErrorInvalidCharacter:    "parse error. invalid character for encoding",
	rpcCodeParseErrorUnsupportedEncoding: "parse error. unsupported encoding",
	rpcCodeParseErrorNotWellFormed:       "parse error. not well formed",

	rpcCodeInternalError:  "internal error",
	rpcCodeInvalidParams:  "invalid method parameters",
	rpcCodeMethodNotFound: "requested method not found",
	rpcCodeInvalidRequest: "invalid request",

	rpcCodeApplicationError: "application error",

	rpcCodeSystemError: "system error",

	rpcCodeTransportError: "transport error",

	rpcCodeAuthenticationFailed: "authentication failed",
	rpcCodeSessionNotFound:      "session not found",
	rpcCodeSessionExpired:       "session expired",
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcRequest struct {
	JsonRpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JsonRpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id,omitempty"`
	Error   interface{}      `json:"error,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
}

type okResult struct {
	Ok      bool        `json:"ok"`
	Message interface{} `json:"msg,omitempty"`
}

func buildRpcError(id *json.RawMessage, code int) *rpcResponse {
	return &rpcResponse{
		JsonRpc: "2.0",
		Id:      id,
		Error: &rpcError{
			Code:    code,
			Message: rpcCodeText[code],
		},
	}
}

func buildRpcResult(id *json.RawMessage, result interface{}) *rpcResponse {
	return &rpcResponse{
		JsonRpc: "2.0",
		Id:      id,
		Result:  result,
	}
}

func sendResponse(w http.ResponseWriter, data interface{}) {
	b, _ := json.Marshal(data)
	_, _ = w.Write(b)
}
