package ulid

import (
	"bufio"
	"encoding/binary"
	"github.com/pixie-sh/errors-go"
	"io"
	"math"
	"math/bits"
	"math/rand"
)

// from https://github.com/oklog/ulid/blob/ulid@v1.3.1/ulid.go
func Monotonic(entropy io.Reader, inc uint64) io.Reader {
	m := monotonic{
		Reader: bufio.NewReader(entropy),
		inc:    inc,
	}

	if m.inc == 0 {
		m.inc = math.MaxUint32
	}

	if rng, ok := entropy.(*rand.Rand); ok {
		m.rng = rng
	}

	return &m
}

type monotonic struct {
	io.Reader
	ms      uint64
	inc     uint64
	entropy uint80
	rand    [8]byte
	rng     *rand.Rand
}

func (m *monotonic) MonotonicRead(ms uint64, entropy []byte) (err error) {
	if !m.entropy.IsZero() && m.ms == ms {
		err = m.increment()
		m.entropy.AppendTo(entropy)
	} else if _, err = io.ReadFull(m.Reader, entropy); err == nil {
		m.ms = ms
		m.entropy.SetBytes(entropy)
	}
	return err
}

// increment the previous entropy number with a random number
// of up to m.inc (inclusive).
func (m *monotonic) increment() error {
	if inc, err := m.random(); err != nil {
		return err
	} else if m.entropy.Add(inc) {
		return errors.New("monotonic overflow").WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}
	return nil
}

// random returns a uniform random value in [1, m.inc), reading entropy
// from m.Reader. When m.inc == 0 || m.inc == 1, it returns 1.
// Adapted from: https://golang.org/pkg/crypto/rand/#Int
func (m *monotonic) random() (inc uint64, err error) {
	if m.inc <= 1 {
		return 1, nil
	}

	// Fast path for using a underlying rand.Rand directly.
	if m.rng != nil {
		// Range: [1, m.inc)
		return 1 + uint64(m.rng.Int63n(int64(m.inc))), nil
	}

	// bitLen is the maximum bit length needed to encode a value < m.inc.
	bitLen := bits.Len64(m.inc)

	// byteLen is the maximum byte length needed to encode a value < m.inc.
	byteLen := uint(bitLen+7) / 8

	// msbitLen is the number of bits in the most significant byte of m.inc-1.
	msbitLen := uint(bitLen % 8)
	if msbitLen == 0 {
		msbitLen = 8
	}

	for inc == 0 || inc >= m.inc {
		if _, err = io.ReadFull(m.Reader, m.rand[:byteLen]); err != nil {
			return 0, err
		}

		// Clear bits in the first byte to increase the probability
		// that the candidate is < m.inc.
		m.rand[0] &= uint8(int(1<<msbitLen) - 1)

		// Convert the read bytes into an uint64 with byteLen
		// Optimized unrolled loop.
		switch byteLen {
		case 1:
			inc = uint64(m.rand[0])
		case 2:
			inc = uint64(binary.LittleEndian.Uint16(m.rand[:2]))
		case 3, 4:
			inc = uint64(binary.LittleEndian.Uint32(m.rand[:4]))
		case 5, 6, 7, 8:
			inc = uint64(binary.LittleEndian.Uint64(m.rand[:8]))
		}
	}

	// Range: [1, m.inc)
	return 1 + inc, nil
}

type uint80 struct {
	Hi uint16
	Lo uint64
}

func (u *uint80) SetBytes(bs []byte) {
	u.Hi = binary.BigEndian.Uint16(bs[:2])
	u.Lo = binary.BigEndian.Uint64(bs[2:])
}

func (u *uint80) AppendTo(bs []byte) {
	binary.BigEndian.PutUint16(bs[:2], u.Hi)
	binary.BigEndian.PutUint64(bs[2:], u.Lo)
}

func (u *uint80) Add(n uint64) (overflow bool) {
	lo, hi := u.Lo, u.Hi
	if u.Lo += n; u.Lo < lo {
		u.Hi++
	}
	return u.Hi < hi
}

func (u uint80) IsZero() bool {
	return u.Hi == 0 && u.Lo == 0
}
