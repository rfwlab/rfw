package host

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	var stream bytes.Buffer
	want := []byte(`{"component":"Counter","payload":{"value":42}}`)
	if err := writeFrame(&stream, want); err != nil {
		t.Fatal(err)
	}
	got, err := readFrame(bufio.NewReader(&stream), 1024)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("frame = %q, want %q", got, want)
	}
}

func TestFrameRejectsPlatformOverflow(t *testing.T) {
	var prefix [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(prefix[:], uint64(^uint(0)>>1)+1)
	if _, err := readFrame(bufio.NewReader(bytes.NewReader(prefix[:n])), 0); err == nil {
		t.Fatal("readFrame accepted a length larger than int")
	}
}
