package service

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"

	"regame-user-service/config"
	"regame-user-service/database"
)

const (
	AuthTypeCode = iota
	AuthTypeSM3
	AuthTypeToken
)

const SM3PasswordLength = 64

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

type UserSession struct {
	Id         string
	CreateTime time.Time
	UpdateTime time.Time
}

type UserService struct {
	sync.Mutex
	cfg      *config.UserServiceConfig
	sessions map[string]*UserSession
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

func hasExpired(t time.Time, duration time.Duration) bool {
	n := time.Now()
	if n.After(t.Add(duration)) {
		return true
	}
	return false
}

func (us *UserService) CleanExpiredSession() {
	us.Lock()
	defer us.Unlock()
	for id, session := range us.sessions {
		if hasExpired(session.UpdateTime, us.cfg.SessionExpiration) {
			delete(us.sessions, id)
		}
	}
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

func (us *UserService) verify(info *loginParams) bool {
	switch info.Type {
	case AuthTypeCode:
		// TODO
	case AuthTypeSM3:
		if len(info.Data) != SM3PasswordLength {
			return false
		}
		password, err := hex.DecodeString(info.Data)
		if err != nil {
			return false
		}
		kv, err := database.TiKVClientGet([]byte(database.PASSWORD_KEY + info.UserName))
		if err == nil {
			if bytes.Equal(kv.V, password) {
				return true
			}
		}
	}

	return false
}

func (us *UserService) rpcLoginHandler(w http.ResponseWriter, req *rpcRequest) {
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

	if !us.verify(reqParams) {
		glog.Errorf("login type %d username %s data %s failed\n",
			reqParams.Type, reqParams.UserName, reqParams.Data)
		res := buildRpcError(req.Id, rpcCodeAuthenticationFailed)
		sendResponse(w, res)
		return
	}

	n := time.Now()
	session := &UserSession{
		Id:         uuid.NewString(),
		CreateTime: n,
		UpdateTime: n,
	}

	us.Lock()
	defer us.Unlock()
	us.sessions[session.Id] = session

	result := loginResult{
		SessionId: session.Id,
		Interval:  time.Duration(us.cfg.KeepAliveDuration.Seconds()),
	}
	res := buildRpcResult(req.Id, result)
	sendResponse(w, res)
	glog.Infof("login type %d username %s data %s ok, session_id %s\n",
		reqParams.Type, reqParams.UserName, reqParams.Data, session.Id)
}

func (us *UserService) rpcKeepAliveHandler(w http.ResponseWriter, req *rpcRequest) {
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

	us.Lock()
	defer us.Unlock()
	session, ok := us.sessions[reqParams.SessionId]
	if !ok {
		glog.Errorf("keepalive session_id %s not found\n", reqParams.SessionId)
		res := buildRpcError(req.Id, rpcCodeSessionNotFound)
		sendResponse(w, res)
		return
	}

	if hasExpired(session.UpdateTime, us.cfg.SessionExpiration) {
		glog.Errorf("keepalive session_id %s expired\n", reqParams.SessionId)
		res := buildRpcError(req.Id, rpcCodeSessionExpired)
		sendResponse(w, res)
		return
	}

	session.UpdateTime = time.Now()
	result := keepAliveResult{
		SessionId: session.Id,
		Interval:  time.Duration(us.cfg.KeepAliveDuration.Seconds()),
	}
	res := buildRpcResult(req.Id, result)
	sendResponse(w, res)
}

func (us *UserService) rpcLogoutHandler(w http.ResponseWriter, req *rpcRequest) {
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

	us.Lock()
	defer us.Unlock()
	delete(us.sessions, reqParams.SessionId)

	w.WriteHeader(http.StatusOK)
	glog.Infof("logout session_id %s logout\n", reqParams.SessionId)
}

func (us *UserService) UserHandler(resWriter http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	req := &rpcRequest{}
	err := dec.Decode(req)
	if err != nil {
		glog.Error("Decode failed: ", err)
		res := buildRpcError(nil, rpcCodeParseError)
		sendResponse(resWriter, res)
		return
	}

	switch req.Method {
	case "login":
		us.rpcLoginHandler(resWriter, req)
	case "keepalive":
		us.rpcKeepAliveHandler(resWriter, req)
	case "logout":
		// notification
		us.rpcLogoutHandler(resWriter, req)
	default:
		res := buildRpcError(nil, rpcCodeMethodNotFound)
		sendResponse(resWriter, res)
	}
}
