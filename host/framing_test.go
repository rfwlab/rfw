package host

import (
	"bufio"
	"bytes"
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
