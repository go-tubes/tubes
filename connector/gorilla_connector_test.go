package tubes_connector_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-tubes/tubes"
	tubes_connector "github.com/go-tubes/tubes/connector"
	tubespb "github.com/go-tubes/tubes/protobuf"
	"github.com/gorilla/websocket"
)

// TestGorillaConnectorFrameType verifies that the gorilla connector writes text
// frames for the JSON codec and binary frames for the protobuf codec, and that
// a full subscribe -> message -> broadcast cycle round-trips over a real
// WebSocket under each codec.
func TestGorillaConnectorFrameType(t *testing.T) {
	cases := []struct {
		name      string
		codec     tubes.Codec
		wantFrame int
	}{
		{"json text frames", tubes.JSONCodec{}, websocket.TextMessage},
		{"protobuf binary frames", tubespb.Codec{}, websocket.BinaryMessage},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
			connector := tubes_connector.NewGorillaConnector(upgrader, func(err *tubes.Error) {
				t.Logf("server error: %s", err.Description)
			})
			system := tubes.New(connector, tubes.WithCodec(tc.codec))
			system.RegisterChannel("/echo/:room", tubes.ChannelHandlers{
				OnMessage: func(c *tubes.Context, m *tubes.Message) {
					c.Broadcast(m.Payload, nil)
				},
			})

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := system.HandleRequest(w, r, map[string]interface{}{}); err != nil {
					t.Errorf("HandleRequest: %v", err)
				}
			}))
			defer server.Close()

			wsURL := strings.Replace(server.URL, "http", "ws", 1)
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("dial: %v", err)
			}
			defer conn.Close()

			clientFrame := websocket.TextMessage
			if tc.codec.Binary() {
				clientFrame = websocket.BinaryMessage
			}

			channelPath := "/echo/demo"
			payload := []byte(`{"text":"hi"}`)

			sub, _ := tc.codec.Marshal(&tubes.Message{Type: tubes.MessageTypeSubscribe, Channel: channelPath})
			if err := conn.WriteMessage(clientFrame, sub); err != nil {
				t.Fatalf("write subscribe: %v", err)
			}
			msg, _ := tc.codec.Marshal(&tubes.Message{Type: tubes.MessageTypeChannelMessage, Channel: channelPath, Payload: payload})
			if err := conn.WriteMessage(clientFrame, msg); err != nil {
				t.Fatalf("write message: %v", err)
			}

			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			frameType, data, err := conn.ReadMessage()
			if err != nil {
				t.Fatalf("read broadcast: %v", err)
			}
			if frameType != tc.wantFrame {
				t.Errorf("broadcast frame type = %d, want %d", frameType, tc.wantFrame)
			}

			var out tubes.Message
			if err := tc.codec.Unmarshal(data, &out); err != nil {
				t.Fatalf("unmarshal broadcast: %v", err)
			}
			if !bytes.Equal(out.Payload, payload) {
				t.Errorf("payload = %s, want %s", out.Payload, payload)
			}
		})
	}
}
