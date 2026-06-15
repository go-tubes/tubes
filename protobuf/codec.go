// Package protobuf provides a binary, Protocol-Buffers-based tubes.Codec.
//
// The message envelope is encoded as the tubes.v1.Envelope defined in
// protobuf/proto/envelope.proto, using the canonical protobuf wire format and
// written as binary WebSocket frames. The application payload is carried
// verbatim in the Envelope.payload bytes field.
//
// Usage:
//
//	import (
//	    "github.com/go-tubes/tubes"
//	    tubespb "github.com/go-tubes/tubes/protobuf"
//	)
//
//	system := tubes.New(connector, tubes.WithCodec(tubespb.Codec{}))
//
// The same codec must be configured on the client (e.g. createProtobufCodec()
// from @go-tubes/tubes-js).
//
// The envelope is encoded by hand with google.golang.org/protobuf/encoding/protowire
// rather than generated code, so the core library carries no generated
// descriptors and stays lean. The wire format is identical to code generated
// from envelope.proto, so it interoperates with protobuf-es on the client.
package protobuf

import (
	"github.com/go-tubes/tubes"
	"google.golang.org/protobuf/encoding/protowire"
)

// Field numbers from envelope.proto (tubes.v1.Envelope).
const (
	fieldType    protowire.Number = 1
	fieldChannel protowire.Number = 2
	fieldPayload protowire.Number = 3
)

// MessageType enum values from envelope.proto.
const (
	typeUnspecified uint64 = 0
	typeSubscribe   uint64 = 1
	typeUnsubscribe uint64 = 2
	typeMessage     uint64 = 3
)

// Codec encodes the go-tubes envelope as a protobuf tubes.v1.Envelope over
// binary WebSocket frames.
type Codec struct{}

// Binary reports that this codec produces binary WebSocket frames.
func (Codec) Binary() bool { return true }

// Marshal encodes m as a tubes.v1.Envelope in protobuf wire format.
func (Codec) Marshal(m *tubes.Message) ([]byte, error) {
	var b []byte
	if t := typeToEnum(m.Type); t != typeUnspecified {
		b = protowire.AppendTag(b, fieldType, protowire.VarintType)
		b = protowire.AppendVarint(b, t)
	}
	if m.Channel != "" {
		b = protowire.AppendTag(b, fieldChannel, protowire.BytesType)
		b = protowire.AppendBytes(b, []byte(m.Channel))
	}
	if len(m.Payload) > 0 {
		b = protowire.AppendTag(b, fieldPayload, protowire.BytesType)
		b = protowire.AppendBytes(b, m.Payload)
	}
	return b, nil
}

// Unmarshal decodes a tubes.v1.Envelope from protobuf wire format into m.
func (Codec) Unmarshal(data []byte, m *tubes.Message) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]

		switch {
		case num == fieldType && typ == protowire.VarintType:
			v, vn := protowire.ConsumeVarint(data)
			if vn < 0 {
				return protowire.ParseError(vn)
			}
			m.Type = enumToType(v)
			data = data[vn:]
		case num == fieldChannel && typ == protowire.BytesType:
			v, vn := protowire.ConsumeBytes(data)
			if vn < 0 {
				return protowire.ParseError(vn)
			}
			m.Channel = string(v)
			data = data[vn:]
		case num == fieldPayload && typ == protowire.BytesType:
			v, vn := protowire.ConsumeBytes(data)
			if vn < 0 {
				return protowire.ParseError(vn)
			}
			// Copy: v aliases the input buffer, which the caller may reuse.
			m.Payload = append([]byte(nil), v...)
			data = data[vn:]
		default:
			// Unknown field: skip it for forward compatibility.
			vn := protowire.ConsumeFieldValue(num, typ, data)
			if vn < 0 {
				return protowire.ParseError(vn)
			}
			data = data[vn:]
		}
	}
	return nil
}

func typeToEnum(t string) uint64 {
	switch t {
	case tubes.MessageTypeSubscribe:
		return typeSubscribe
	case tubes.MessageTypeUnsubscribe:
		return typeUnsubscribe
	case tubes.MessageTypeChannelMessage:
		return typeMessage
	default:
		return typeUnspecified
	}
}

func enumToType(v uint64) string {
	switch v {
	case typeSubscribe:
		return tubes.MessageTypeSubscribe
	case typeUnsubscribe:
		return tubes.MessageTypeUnsubscribe
	case typeMessage:
		return tubes.MessageTypeChannelMessage
	default:
		return ""
	}
}
