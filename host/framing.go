package host

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

func writeFrame(w io.Writer, payload []byte) error {
	var prefix [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(prefix[:], uint64(len(payload)))
	if err := writeAll(w, prefix[:n]); err != nil {
		return err
	}
	return writeAll(w, payload)
}

func readFrame(r *bufio.Reader, maximum int) ([]byte, error) {
	size, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}
	if maximum > 0 && size > uint64(maximum) {
		return nil, fmt.Errorf("host: frame length %d exceeds %d", size, maximum)
	}
	payload := make([]byte, int(size))
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeAll(w io.Writer, payload []byte) error {
	for len(payload) > 0 {
		n, err := w.Write(payload)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		payload = payload[n:]
	}
	return nil
}
