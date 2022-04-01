# User Service for Regame implemented in Node.js

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
