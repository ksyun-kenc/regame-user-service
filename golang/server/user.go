package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"

	"regame-user-service/config"
)

const (
	rpcCodeAuthenticationFailed = -101
	rpcCodeSessionNotFound      = -201
	rpcCodeSessionExpired       = -202

	rpcCodeParseError     = -32700
	rpcCodeInvalidRequest = -32600
	rpcCodeMethodNotFound = -32601
)

var rpcCodeText = map[int]string{
	rpcCodeAuthenticationFailed: "authentication failed",
	rpcCodeSessionNotFound:      "session not found",
	rpcCodeSessionExpired:       "session expired",

	rpcCodeParseError:     "parse error",
	rpcCodeInvalidRequest: "invalid request",
	rpcCodeMethodNotFound: "method not found",
}

type Session struct {
	Id         string
	CreateTime time.Time
	UpdateTime time.Time
}

type rpcRequest struct {
	JsonRpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JsonRpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id,omitempty"`
	Error   interface{}      `json:"error,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
}

type loginParams struct {
	Version  int    `json:"version"`
	Type     int    `json:"type"`
	UserName string `json:"username"`
	Data     string `json:"data"`
}

type loginResult struct {
	SessionId string        `json:"session_id"`
	Interval  time.Duration `json:"interval"`
}

type keepAliveParams struct {
	SessionId string `json:"session_id"`
}

type keepAliveResult struct {
	SessionId string        `json:"session_id"`
	Interval  time.Duration `json:"interval"`
}

type logoutParams struct {
	SessionId string `json:"session_id"`
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

func (s *Server) auth(info *loginParams) bool {
	if info.Type == config.AuthTypeSM3 {
		for _, sm3Auth := range s.cfg.SM3AuthCfg {
			if sm3Auth.UserName == info.UserName && sm3Auth.Data == info.Data {
				return true
			}
		}
	} else {
		// TODO: Dynamic( AuthTypeCode、AuthTypeToken ) verification is required
	}

	return false
}

func (s *Server) rpcLoginHandler(w http.ResponseWriter, req *rpcRequest) {
	if req.Id == nil || req.Params == nil {
		res := buildRpcError(nil, rpcCodeInvalidRequest)
		sendResponse(w, res)
		return
	}

	reqParams := &loginParams{}
	err := json.Unmarshal(*req.Params, reqParams)
	if err != nil {
		res := buildRpcError(req.Id, rpcCodeInvalidRequest)
		sendResponse(w, res)
		return
	}

	if !s.auth(reqParams) {
		glog.Errorf("login type %d username %s data %s failed\n",
			reqParams.Type, reqParams.UserName, reqParams.Data)
		res := buildRpcError(req.Id, rpcCodeAuthenticationFailed)
		sendResponse(w, res)
		return
	}

	n := time.Now()
	session := &Session{
		Id:         uuid.NewString(),
		CreateTime: n,
		UpdateTime: n,
	}

	s.Lock()
	defer s.Unlock()
	s.sessions[session.Id] = session

	result := loginResult{
		SessionId: session.Id,
		Interval:  time.Duration(s.cfg.KeepAliveDuration.Seconds()),
	}
	res := buildRpcResult(req.Id, result)
	sendResponse(w, res)
	glog.Infof("login type %d username %s data %s ok, session_id %s\n",
		reqParams.Type, reqParams.UserName, reqParams.Data, session.Id)
}

func (s *Server) rpcKeepAliveHandler(w http.ResponseWriter, req *rpcRequest) {
	if req.Id == nil || req.Params == nil {
		res := buildRpcError(nil, rpcCodeInvalidRequest)
		sendResponse(w, res)
		return
	}

	reqParams := &keepAliveParams{}
	err := json.Unmarshal(*req.Params, reqParams)
	if err != nil {
		res := buildRpcError(req.Id, rpcCodeInvalidRequest)
		sendResponse(w, res)
		return
	}

	s.Lock()
	defer s.Unlock()
	session, ok := s.sessions[reqParams.SessionId]
	if !ok {
		glog.Errorf("keepalive session_id %s not found\n", reqParams.SessionId)
		res := buildRpcError(req.Id, rpcCodeSessionNotFound)
		sendResponse(w, res)
		return
	}

	if isTimeExpired(session.UpdateTime, s.cfg.ExpiredDuration) {
		glog.Errorf("keepalive session_id %s expired\n", reqParams.SessionId)
		res := buildRpcError(req.Id, rpcCodeSessionExpired)
		sendResponse(w, res)
		return
	}

	session.UpdateTime = time.Now()
	result := keepAliveResult{
		SessionId: session.Id,
		Interval:  time.Duration(s.cfg.KeepAliveDuration.Seconds()),
	}
	res := buildRpcResult(req.Id, result)
	sendResponse(w, res)
}

func (s *Server) rpcLogoutHandler(w http.ResponseWriter, req *rpcRequest) {
	// jsonrpc notification, Id is nil
	if req.Id != nil || req.Params == nil {
		res := buildRpcError(nil, rpcCodeInvalidRequest)
		sendResponse(w, res)
		return
	}

	reqParams := &logoutParams{}
	err := json.Unmarshal(*req.Params, reqParams)
	if err != nil {
		res := buildRpcError(nil, rpcCodeInvalidRequest)
		sendResponse(w, res)
		return
	}

	s.Lock()
	defer s.Unlock()
	delete(s.sessions, reqParams.SessionId)

	w.WriteHeader(http.StatusOK)
	glog.Infof("logout session_id %s logout\n", reqParams.SessionId)
}

func (s *Server) UserHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	req := &rpcRequest{}
	err := dec.Decode(req)
	if err != nil {
		res := buildRpcError(nil, rpcCodeParseError)
		sendResponse(w, res)
		return
	}

	switch req.Method {
	case "login":
		s.rpcLoginHandler(w, req)
	case "keepalive":
		s.rpcKeepAliveHandler(w, req)
	case "logout":
		// notification
		s.rpcLogoutHandler(w, req)
	default:
		res := buildRpcError(nil, rpcCodeMethodNotFound)
		sendResponse(w, res)
	}
}
