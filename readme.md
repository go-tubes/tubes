<h1 align="center">
  <img src="https://raw.githubusercontent.com/go-tubes/tubes/images/logo.png"><br>
  Tubes - WebSocket Channel Management
</h1>


[![Run Tests](https://github.com/go-tubes/tubes/actions/workflows/run-tests.yml/badge.svg?branch=main)](https://github.com/go-tubes/tubes/actions/workflows/run-tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-tubes/tubes)](https://goreportcard.com/report/github.com/go-tubes/tubes)
[![codecov](https://codecov.io/gh/go-tubes/tubes/branch/main/graph/badge.svg?token=9VA6CYDXAZ)](https://codecov.io/gh/go-tubes/tubes)

tubes is a flexible package for managing Pub-Sub over WebSockets in Go. It offers a rest-style syntax and easily integrates with various websocket and http frameworks.

## Installation

### Installing the main library

1. Get the `tubes` package using the following command:

```shell
go get github.com/go-tubes/tubes
```

## Client Libraries

For client-side integration, you can use one of the following client libraries:

| Language | URL                                              |
| -------- |--------------------------------------------------|
| JavaScript | [GitHub](https://github.com/go-tubes/tubes-js)   |
| Dart | [GitHub](https://github.com/go-tubes/tubes-dart) |

## Connectors

For server-side integration with WebSocket libraries, you can use one of the following connectors:

| WebSocket Library | Status |
| ----------------- |--------|
| Gorilla WebSocket | ✅      |
| Melody | ✅      |

## Getting Started

1. Create a new TubeSystem

```go
tubeSystem := tubes.New(tubes_connector.NewGorillaConnector(
  websocket.Upgrader{},
  func(err *tubes.Error) {
    println(err.Description)
  },
))
```

2. Register Channels

```go
tubeSystem.RegisterChannel("/stream/:streamId", tubes.ChannelHandlers{
  OnSubscribe: func(s *tubes.Context) {
    println("Client joined: " + s.FullPath)
  },
  OnMessage: func(s *tubes.Context, message *tubes.Message) {
    println("New Message on " + s.FullPath + ": " + string(message.Payload))
  },
  OnUnsubscribe: func(s *tubes.Context) {
    println("Client left: " + s.FullPath)
  },
})
```

3. Provide a connect route

```go
r.GET("/connect", func(c *gin.Context) {
  properties := make(map[string]interface{}, 1)
  properties["ctx"] = c

  if err := tubeSystem.HandleRequest(c.Writer, c.Request, properties); err != nil {
    println("Something went wrong while handling a Socket request")
  }
})
```

4. Connect from a frontend lib
```javascript
const client = new TubesClient({ url: socketUrl, debugging: true })
client.subscribeChannel("test", console.log);
client.send("test", { payload: { foo: "bar" } })
```

## Examples

To get a quick overview of how to use tubes, check out the `examples` folder.
