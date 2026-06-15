package tubes

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONCodecBinary(t *testing.T) {
	if (JSONCodec{}).Binary() {
		t.Errorf("JSONCodec.Binary() = true, want false")
	}
}

func TestJSONCodecRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		msgType string
		channel string
		payload json.RawMessage
	}{
		{"subscribe no payload", MessageTypeSubscribe, "example/path/a", nil},
		{"unsubscribe no payload", MessageTypeUnsubscribe, "example/path/b", nil},
		{"message json object", MessageTypeChannelMessage, "example/path/c", json.RawMessage(`{"name":"Jon","admin":false}`)},
		{"message json string", MessageTypeChannelMessage, "example/path/d", json.RawMessage(`"hello"`)},
	}

	codec := JSONCodec{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &Message{Type: tc.msgType, Channel: tc.channel, Payload: tc.payload}
			data, err := codec.Marshal(in)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var out Message
			if err := codec.Unmarshal(data, &out); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if out.Type != tc.msgType {
				t.Errorf("Type = %q, want %q", out.Type, tc.msgType)
			}
			if out.Channel != tc.channel {
				t.Errorf("Channel = %q, want %q", out.Channel, tc.channel)
			}
			if tc.payload != nil && !bytes.Equal(out.Payload, tc.payload) {
				t.Errorf("Payload = %s, want %s", out.Payload, tc.payload)
			}
		})
	}
}

func TestJSONCodecUnmarshalError(t *testing.T) {
	var out Message
	if err := (JSONCodec{}).Unmarshal([]byte{0x01, 0x02, 0x03}, &out); err == nil {
		t.Errorf("Unmarshal(garbage) = nil error, want error")
	}
}

// TestNewDefaultsToJSON verifies the default codec is JSONCodec (text frames)
// when WithCodec is not supplied.
func TestNewDefaultsToJSON(t *testing.T) {
	fakeConnector, _ := NewFakeConnector(func(*Error) {})
	system := New(fakeConnector)

	if _, ok := system.codec.(JSONCodec); !ok {
		t.Errorf("default codec = %T, want JSONCodec", system.codec)
	}
	if fakeConnector.Binary() {
		t.Errorf("connector.Binary() = true for default JSON codec, want false")
	}
}

// stubBinaryCodec is a JSON-on-the-wire codec that reports Binary()==true, used
// to verify WithCodec threads the binary flag to the connector without pulling
// the protobuf subpackage into this (internal) test package.
type stubBinaryCodec struct{ JSONCodec }

func (stubBinaryCodec) Binary() bool { return true }

func TestWithCodecSetsConnectorBinary(t *testing.T) {
	fakeConnector, _ := NewFakeConnector(func(*Error) {})
	system := New(fakeConnector, WithCodec(stubBinaryCodec{}))

	if _, ok := system.codec.(stubBinaryCodec); !ok {
		t.Errorf("codec = %T, want stubBinaryCodec", system.codec)
	}
	if !fakeConnector.Binary() {
		t.Errorf("connector.Binary() = false, want true for a binary codec")
	}
}

func TestWithCodecNilKeepsDefault(t *testing.T) {
	fakeConnector, _ := NewFakeConnector(func(*Error) {})
	system := New(fakeConnector, WithCodec(nil))

	if _, ok := system.codec.(JSONCodec); !ok {
		t.Errorf("codec = %T after WithCodec(nil), want JSONCodec", system.codec)
	}
}
