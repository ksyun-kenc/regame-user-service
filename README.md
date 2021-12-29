# User Service for Regame

`regame-user-service` is a JSON-RPC service for [鎏光/regame](https://github.com/ksyun-kenc/liuguang) to maintain user state.

## Install

```sh
npm install
```

## Run

```sh
node index.js
```

## Test

```sh
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","id":0,"method":"login","params":{"version":0,"username":"UMU","type":0,"data":"123456"}}' http://127.0.0.1:8545/
```
