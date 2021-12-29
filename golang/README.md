# User Service for Regame

`regame-user-service` is a JSON-RPC over http service for [鎏光/regame](https://github.com/ksyun-kenc/liuguang) to maintain user state.

## build

```sh
go build -ldflags "-s -w" -o service cmd/main.go
```

## Run

```sh
./service config/config.json
```

## Test

- login
```sh
# run
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","id":0,"method":"login","params":{"version":0,"username":"UMU","type":0,"data":"123456"}}' http://127.0.0.1:8545/user
# response
{"jsonrpc":"2.0","id":0,"result":{"session_id":"ae692959-32e4-4f88-9c0f-13c25380baee","interval":10}}
```

- keepalive
```sh
# run
 curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","id":0,"method":"keepalive", "params":{"session_id": "ae692959-32e4-4f88-9c0f-13c25380baee"}}' http://127.0.0.1:8545/user
 # response
 {"jsonrpc":"2.0","id":0,"result":{"session_id":"ae692959-32e4-4f88-9c0f-13c25380baee","interval":10}}
 ```

- logout
```sh
# run
 curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"logout", "params":{"session_id": "ae692959-32e4-4f88-9c0f-13c25380baee"}}' http://127.0.0.1:8545/user
```