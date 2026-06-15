package protobuf

import (
	"bytes"
	"testing"

	"github.com/go-tubes/tubes"
)

func TestCodecBinary(t *testing.T) {
	if !(Codec{}).Binary() {
		t.Errorf("Codec.Binary() = false, want true")
	}
}

func TestCodecRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		msgType string
		channel string
		payload []byte
	}{
		{"subscribe", tubes.MessageTypeSubscribe, "example/path/a", nil},
		{"unsubscribe", tubes.MessageTypeUnsubscribe, "example/path/b", nil},
		{"message empty payload", tubes.MessageTypeChannelMessage, "example/path/c", []byte{}},
		{"message text payload", tubes.MessageTypeChannelMessage, "/echo/demo", []byte("hello world")},
		{"message non-utf8 payload", tubes.MessageTypeChannelMessage, "/echo/demo", []byte{0x00, 0xff, 0xfe, 0x10, 0x80}},
	}

	codec := Codec{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &tubes.Message{Type: tc.msgType, Channel: tc.channel, Payload: tc.payload}
			data, err := codec.Marshal(in)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var out tubes.Message
			if err := codec.Unmarshal(data, &out); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if out.Type != tc.msgType {
				t.Errorf("Type = %q, want %q", out.Type, tc.msgType)
			}
			if out.Channel != tc.channel {
				t.Errorf("Channel = %q, want %q", out.Channel, tc.channel)
			}
			if len(tc.payload) > 0 && !bytes.Equal(out.Payload, tc.payload) {
				t.Errorf("Payload = %v, want %v", out.Payload, tc.payload)
			}
		})
	}
}

// TestCodecEnumMapping checks the MessageType string <-> enum mapping in both
// directions, including the unspecified/unknown cases.
func TestCodecEnumMapping(t *testing.T) {
	known := []struct {
		str  string
		enum uint64
	}{
		{tubes.MessageTypeSubscribe, typeSubscribe},
		{tubes.MessageTypeUnsubscribe, typeUnsubscribe},
		{tubes.MessageTypeChannelMessage, typeMessage},
	}
	for _, k := range known {
		if got := typeToEnum(k.str); got != k.enum {
			t.Errorf("typeToEnum(%q) = %d, want %d", k.str, got, k.enum)
		}
		if got := enumToType(k.enum); got != k.str {
			t.Errorf("enumToType(%d) = %q, want %q", k.enum, got, k.str)
		}
	}

	if got := typeToEnum("nonsense"); got != typeUnspecified {
		t.Errorf("typeToEnum(unknown) = %d, want %d", got, typeUnspecified)
	}
	if got := enumToType(99); got != "" {
		t.Errorf("enumToType(99) = %q, want \"\"", got)
	}
}

// TestCodecGoldenWire locks the exact protobuf wire bytes so the encoding stays
// compatible with code generated from envelope.proto (e.g. protobuf-es on the
// client). Fields are emitted in order: type(1, varint), channel(2, bytes),
// payload(3, bytes).
func TestCodecGoldenWire(t *testing.T) {
	in := &tubes.Message{
		Type:    tubes.MessageTypeChannelMessage, // enum 3
		Channel: "ab",
		Payload: []byte{0x01, 0x02},
	}
	want := []byte{
		0x08, 0x03, // field 1 (type), varint 3
		0x12, 0x02, 'a', 'b', // field 2 (channel), len 2
		0x1a, 0x02, 0x01, 0x02, // field 3 (payload), len 2
	}
	got, err := (Codec{}).Marshal(in)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("Marshal() = % x, want % x", got, want)
	}
}

func TestCodecUnspecifiedTypeOmitted(t *testing.T) {
	// An unknown type maps to MESSAGE_TYPE_UNSPECIFIED (0) and is omitted from
	// the wire (proto3 default). Decoding yields an empty type string.
	in := &tubes.Message{Type: "weird", Channel: "x", Payload: nil}
	data, err := (Codec{}).Marshal(in)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var out tubes.Message
	if err := (Codec{}).Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if out.Type != "" {
		t.Errorf("Type = %q, want empty string", out.Type)
	}
}

func TestCodecUnmarshalError(t *testing.T) {
	// A truncated length-delimited field (claims 5 bytes, supplies none) must
	// produce an error rather than panic.
	bad := []byte{0x12, 0x05}
	var out tubes.Message
	if err := (Codec{}).Unmarshal(bad, &out); err == nil {
		t.Errorf("Unmarshal(truncated) = nil error, want error")
	}
}

func TestCodecUnknownFieldSkipped(t *testing.T) {
	// Forward compatibility: an unknown field (number 5, varint) is skipped and
	// known fields still decode.
	data := []byte{
		0x28, 0x07, // field 5, varint 7 (unknown)
		0x12, 0x01, 'z', // field 2 (channel) = "z"
	}
	var out tubes.Message
	if err := (Codec{}).Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if out.Channel != "z" {
		t.Errorf("Channel = %q, want %q", out.Channel, "z")
	}
}
