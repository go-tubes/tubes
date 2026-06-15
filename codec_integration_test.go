package tubes_test

import (
	"bytes"
	"testing"

	"github.com/go-tubes/tubes"
	tubespb "github.com/go-tubes/tubes/protobuf"
)

// TestCodecEndToEnd drives a full subscribe -> message -> broadcast -> receive
// cycle through the fake connector under each codec, asserting the application
// payload survives the envelope round-trip on the wire.
func TestCodecEndToEnd(t *testing.T) {
	codecs := []struct {
		name  string
		codec tubes.Codec
	}{
		{"json", tubes.JSONCodec{}},
		{"protobuf", tubespb.Codec{}},
	}

	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			channelPath := "/echo/demo"
			payload := []byte(`{"text":"hello"}`)

			fakeConnector, fakeSocket := tubes.NewFakeConnector(func(err *tubes.Error) {
				t.Errorf("unexpected error: %s", err.Description)
			})
			system := tubes.New(fakeConnector, tubes.WithCodec(tc.codec))

			system.RegisterChannel("/echo/:room", tubes.ChannelHandlers{
				OnMessage: func(c *tubes.Context, m *tubes.Message) {
					c.Broadcast(m.Payload, nil)
				},
			})

			var received [][]byte
			client := fakeSocket.NewClientConnects(func(msg []byte) {
				received = append(received, msg)
			})

			subBytes, err := tc.codec.Marshal(&tubes.Message{Type: tubes.MessageTypeSubscribe, Channel: channelPath})
			if err != nil {
				t.Fatalf("marshal subscribe: %v", err)
			}
			client.Send(subBytes)

			msgBytes, err := tc.codec.Marshal(&tubes.Message{Type: tubes.MessageTypeChannelMessage, Channel: channelPath, Payload: payload})
			if err != nil {
				t.Fatalf("marshal message: %v", err)
			}
			client.Send(msgBytes)

			if len(received) != 1 {
				t.Fatalf("received %d broadcasts, want 1", len(received))
			}

			var out tubes.Message
			if err := tc.codec.Unmarshal(received[0], &out); err != nil {
				t.Fatalf("unmarshal broadcast: %v", err)
			}
			if out.Channel != channelPath {
				t.Errorf("broadcast channel = %q, want %q", out.Channel, channelPath)
			}
			if !bytes.Equal(out.Payload, payload) {
				t.Errorf("broadcast payload = %s, want %s", out.Payload, payload)
			}
		})
	}
}
