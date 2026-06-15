package tubes

import "encoding/json"

// Codec controls how the message envelope (Message) is serialized on the wire.
//
// Binary reports whether the resulting frames should be written as binary
// WebSocket frames (true) or text frames (false). The connector consults this
// to pick the correct frame opcode.
//
// The application payload (Message.Payload) is treated as opaque bytes by every
// codec; only the envelope framing differs.
type Codec interface {
	Marshal(*Message) ([]byte, error)
	Unmarshal([]byte, *Message) error
	Binary() bool
}

// JSONCodec is the default Codec. It encodes the envelope as JSON over text
// frames, preserving the historical go-tubes wire format. It is the zero-value
// behaviour: a TubeSystem created without WithCodec uses JSONCodec.
type JSONCodec struct{}

func (JSONCodec) Marshal(m *Message) ([]byte, error)   { return json.Marshal(m) }
func (JSONCodec) Unmarshal(b []byte, m *Message) error { return json.Unmarshal(b, m) }
func (JSONCodec) Binary() bool                         { return false }

// Option configures a TubeSystem at construction time.
type Option func(*TubeSystem)

// WithCodec sets the envelope Codec used for every message on the connection.
// When omitted, the TubeSystem defaults to JSONCodec (text frames). The same
// codec must be configured on both the server and the client.
func WithCodec(c Codec) Option {
	return func(r *TubeSystem) {
		if c != nil {
			r.codec = c
		}
	}
}
