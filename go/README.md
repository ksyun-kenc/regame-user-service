# User Service for Regame implemented in Go

## Build

Windows:

```Powershell
go build -ldflags "-s -w" -o regame_user_service.exe .\cmd\main.go
go build -ldflags "-s -w" -o add_user.exe .\cmd\add_user.go
```

macOS or Linux:

```sh
go build -ldflags "-s -w" -o regame_user_service cmd/main.go
```

## Prepare database

1. Start a PD on your server.

Detail: <https://github.com/tikv/tikv#deploy-a-playground-with-binary>

2. Run user_manager.

Windows:

```Powershell
.\user_manager.exe -pd tikv://$PD_ADDR:2379 -op add UMU 207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb
.\user_manager.exe -pd tikv://$PD_ADDR:2379 -op list
```

## Run

Windows:

```Powershell
.\regame_user_service.exe .\config\config.json
```

macOS or Linux:

```sh
./regame_user_service config/config.json
```

## Test

On Windows, you should replace `"` with `\"` in `--data` parameter. eg:

```Powershell
curl -X POST -H "Content-Type: application/json" --data '{\"jsonrpc\":\"2.0\",\"id\":0,\"method\":\"login\",\"params\":{\"version\":0,\"username\":\"UMU\",\"type\":1,\"data\":\"207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb\"}}' http://127.0.0.1:8545/user
```

- login

```sh
# request
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","id":0,"method":"login","params":{"version":0,"username":"UMU","type":1,"data":"207cf410532f92a47dee245ce9b11ff71f578ebd763eb3bbea44ebd043d018fb"}}' http://127.0.0.1:8545/user

# response
{"jsonrpc":"2.0","id":0,"result":{"session_id":"ae692959-32e4-4f88-9c0f-13c25380baee","interval":10}}
```

- keepalive

```sh
# request
 curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","id":0,"method":"keepalive", "params":{"session_id": "ae692959-32e4-4f88-9c0f-13c25380baee"}}' http://127.0.0.1:8545/user

 # response
 {"jsonrpc":"2.0","id":0,"result":{"session_id":"ae692959-32e4-4f88-9c0f-13c25380baee","interval":10}}
 ```

- logout
```sh
# request
 curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"logout", "params":{"session_id": "ae692959-32e4-4f88-9c0f-13c25380baee"}}' http://127.0.0.1:8545/user
```
