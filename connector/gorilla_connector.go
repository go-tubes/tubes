package tubes_connector

import (
	"fmt"
	"github.com/go-tubes/tubes"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

func NewGorillaConnector(upgrader websocket.Upgrader, errorHandler tubes.ErrorHandlerFunc) *tubes.Connector {
	var connector *tubes.Connector
	connector = tubes.NewConnector(
		func(writer http.ResponseWriter, request *http.Request, properties map[string]interface{}) error {
			mutex := sync.Mutex{}

			conn, err := upgrader.Upgrade(writer, request, nil)
			if err != nil {
				return err
			}

			client := connector.Join(
				func(message []byte) error {
					mutex.Lock()
					defer mutex.Unlock()
					return conn.WriteMessage(websocket.TextMessage, message)
				},
				properties,
			)

			go func() {
				defer (func() {
					err := conn.Close()
					if err != nil {
						fmt.Println("Error during closing connection:", err)
					}
					connector.Leave(client.Id)
				})()

				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						switch err.(type) {
						case *websocket.CloseError:
							// omit ual websocket close errors
						default:
							fmt.Println("Error during message reading:", err)
						}
						break
					}
					connector.Message(client.Id, message)
				}
			}()

			return nil
		},
		errorHandler,
	)

	return connector
}
