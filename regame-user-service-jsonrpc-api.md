# User Service for Regame

## API Specification

JSONRPC over HTTP

- Method: POST
- Body: JSONRPC object

## **1. login**

#### Request

```JSON
{
  "jsonrpc": "2.0",
  "id": 0,
  "method": "login",
  "params": {
    "version": 0,
    "type": 2,
    "username": "UMU",
    "data": "4805e236-f49c-43fe-926c-f9b7c9de534a"
  }
}
```

**version**: Currently 0

**type**:

- 0: code
- 1: SM3
- 2: Token

#### Response

##### Succeeded

```JSON
{
  "jsonrpc": "2.0",
  "id": 0,
  "result": {
    "interval": 900,
    "session_id": "4805e236-f49c-43fe-926c-f9b7c9de534a"
  }
}
```

- **interval**: unit is second

- **session_id**: UUID

##### Failed

```JSON
{
  "jsonrpc": "2.0",
  "id": 0,
  "error": {
    "code": -101,
    "message": "authentication failed"
  }
}
```

**code**:

-101: authentication failed

## **2. keepalive**

#### Request

```JSON
{
  "jsonrpc": "2.0",
  "id": 0,
  "method": "keepalive",
  "params": {
    "session_id": "4805e236-f49c-43fe-926c-f9b7c9de534a"
  }
}
```

#### Response

##### Succeeded

The same as `login`.

##### Failed

```JSON
{
  "jsonrpc": "2.0",
  "id": 0,
  "error": {
    "code": -201,
    "message": "session not found"
  }
}
```

**code**:

- 201: session not found
- 202: session expired

## **3. logout**

#### Request

JSONRPC notification

```JSON
{
  "jsonrpc": "2.0",
  "method": "logout",
  "params": {
    "session_id": "4805e236-f49c-43fe-926c-f9b7c9de534a"
  }
}
```
