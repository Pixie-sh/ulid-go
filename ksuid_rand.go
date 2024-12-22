package ulid

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"io"
	"math/rand"
	"sync"
)

func initRandReader() (r io.Reader, err error) {
	var seed int64

	if _, err = io.ReadFull(cryptoRand.Reader, seedBytes[:]); err != nil {
		return
	}

	seed = int64(binary.LittleEndian.Uint64(seedBytes[:]))

	r = &randSourceReader{
		source: rand.NewSource(seed).(rand.Source64),
	}
	return
}

type randSourceReader struct {
	mu     sync.Mutex
	source rand.Source64
}

func (r *randSourceReader) Read(b []byte) (int, error) {
	// optimized for generating 16 bytes payloads
	binary.LittleEndian.PutUint64(b[:8], r.source.Uint64())
	binary.LittleEndian.PutUint64(b[8:], r.source.Uint64())
	return 16, nil
}
