package splithttp_test

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/xtls/xray-core/transport/internet/splithttp"
)

func TestSeqEncodingDecimal(t *testing.T) {
	config := &Config{}

	encoded := config.EncodeSeq("session", 42)
	if encoded != "42" {
		t.Fatalf("unexpected decimal seq: %q", encoded)
	}

	decoded, err := config.DecodeSeq("session", encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != 42 {
		t.Fatalf("unexpected decoded seq: %d", decoded)
	}
}

func TestSeqEncodingMaskedHexRoundTrip(t *testing.T) {
	config := &Config{SeqEncoding: SeqEncodingMaskedHex}
	values := []uint64{0, 1, 2, 15, 16, 255, 256, 1<<32 - 1, 1 << 32, ^uint64(0)}

	for _, value := range values {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			encoded := config.EncodeSeq("session-a", value)
			if len(encoded) != 16 {
				t.Fatalf("unexpected encoded seq length: %q", encoded)
			}
			if _, err := strconv.ParseUint(encoded, 16, 64); err != nil {
				t.Fatalf("encoded seq is not hex: %q", encoded)
			}

			decoded, err := config.DecodeSeq("session-a", encoded)
			if err != nil {
				t.Fatal(err)
			}
			if decoded != value {
				t.Fatalf("unexpected decoded seq: got %d, want %d", decoded, value)
			}
		})
	}
}

func TestSeqEncodingMaskedHexUsesSessionID(t *testing.T) {
	config := &Config{SeqEncoding: SeqEncodingMaskedHex}

	a := config.EncodeSeq("session-a", 0)
	b := config.EncodeSeq("session-b", 0)
	if a == b {
		t.Fatalf("masked seq should depend on session ID: %q", a)
	}
}

func TestSeqEncodingMaskedHexDoesNotExposeSmallCounters(t *testing.T) {
	config := &Config{SeqEncoding: SeqEncodingMaskedHex}

	for seq := uint64(0); seq < 4; seq++ {
		encoded := config.EncodeSeq("session-a", seq)
		if encoded == strconv.FormatUint(seq, 10) || encoded == fmt.Sprintf("%016x", seq) {
			t.Fatalf("masked seq exposes counter %d as %q", seq, encoded)
		}
	}
}

func TestSeqEncodingMaskedHexRejectsInvalidValues(t *testing.T) {
	config := &Config{SeqEncoding: SeqEncodingMaskedHex}

	if _, err := config.DecodeSeq("session", "1"); err == nil {
		t.Fatal("expected error for short masked seq")
	}
	if _, err := config.DecodeSeq("session", "not-a-hex-value!"); err == nil {
		t.Fatal("expected error for invalid masked seq")
	}
}
