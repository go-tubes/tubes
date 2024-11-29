package main

import (
	"github.com/go-tubes/tubes"
	tubes_connector "github.com/go-tubes/tubes/connector"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	r := gin.Default()

	r.Static("js/", "html/node_modules/go-tubes-client/dist/")
	r.LoadHTMLGlob("html/*.html")

	tubeSystem := tubes.New(tubes_connector.NewGorillaConnector(
		websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		func(err *tubes.Error) {
			println(err.Description)
		},
	))

	tubeSystem.RegisterChannel("echo", tubes.ChannelHandlers{
		OnSubscribe: func(s *tubes.Context) {
			println("Client joined: " + s.FullPath)
		},
		OnMessage: func(s *tubes.Context, message *tubes.Message) {
			println("New Message on " + s.FullPath + ": " + string(message.Payload))
			s.Send(message.Payload)
		},
		OnUnsubscribe: func(s *tubes.Context) {
			println("Client left: " + s.FullPath)
		},
	})

	r.Use(func(c *gin.Context) {
		c.Set("tubeSystem", tubeSystem)
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"socketUrl": "ws://localhost:" + port + "/connect",
		})
	})

	r.GET("/connect", func(c *gin.Context) {
		properties := make(map[string]interface{}, 1)
		properties["ctx"] = c

		if err := tubeSystem.HandleRequest(c.Writer, c.Request, properties); err != nil {
			println("Something went wrong while handling a Socket request")
		}
	})

	if err := r.Run(":" + port); err != nil {
		panic("Failed to start server")
	}
}
