package ulid

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"io"
	"time"
	"unsafe"

	"github.com/pixie-sh/errors-go"
)

// ULID inspired on many open source projects, thank you all!
// github.com/google/uuid
// github.com/matoous/go-nanoid/v2
// github.com/oklog/ulid
// github.com/RobThree/NUlid
// github.com/segmentio/ksuid

type ULID [16]byte

func New(customEntropy ...io.Reader) (ULID, error) {
	var (
		id       = EmptyUID
		now      = time.Now()
		entropy  = defaultEntropy
		err      error
	)

	if len(customEntropy) > 0 && customEntropy[0] != nil {
		entropy = customEntropy[0]
	}

	if _, err = entropy.Read(id[6:]); err != nil {
		return id, err
	}

	if err = id.setTime(now); err != nil {
		return id, err
	}

	return id, nil
}

func MustNew(customEntropy ...io.Reader) ULID {
	id, err := New(customEntropy...)
	if err != nil {
		panic(err)
	}

	return id
}

func UnmarshalString(s string) (ULID, error) {
	id := ULID{}

	if err := id.UnmarshalText([]byte(s)); err != nil {
		return EmptyUID, err
	}

	return id, nil
}

func UnmarshalBytes(b []byte) (ULID, error) {
	id := ULID{}

	if err := id.UnmarshalBinary(b); err != nil {
		return EmptyUID, err
	}

	return id, nil
}

func UnmarshalUint64(num uint64) (ULID, error) {
	var (
		out = make([]byte, 16)
		id  = ULID{}
	)

	copy(out, leftPad[:])

	out[7] = byte(num >> 63)
	out[8] = byte(num >> 56)
	out[9] = byte(num >> 48)
	out[10] = byte(num >> 40)
	out[11] = byte(num >> 32)
	out[12] = byte(num >> 24)
	out[13] = byte(num >> 16)
	out[14] = byte(num >> 8)
	out[15] = byte(num)

	if err := id.UnmarshalBinary(out); err != nil {
		return EmptyUID, err
	}

	return id, nil
}

func UnmarshalUUID(s string) (ULID, error) {
	id := ULID{}

	if len(s) != uuidEncodedSize {
		return id, errors.New("invalid data size len(%d)", len(s)).WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}

	_, err := hex.Decode(id[0:4], []byte(s[0:8]))
	if err != nil {
		return [16]byte{}, err
	}

	_, err = hex.Decode(id[4:6], []byte(s[9:13]))
	if err != nil {
		return [16]byte{}, err
	}

	_, err = hex.Decode(id[6:8], []byte(s[14:18]))
	if err != nil {
		return [16]byte{}, err
	}

	_, err = hex.Decode(id[8:10], []byte(s[19:23]))
	if err != nil {
		return [16]byte{}, err
	}

	_, err = hex.Decode(id[10:], []byte(s[24:]))
	if err != nil {
		return [16]byte{}, err
	}

	return id, nil
}

func (id ULID) MarshalBinary() ([]byte, error) {
	dst := make([]byte, len(id))

	copy(dst, id[:])

	return dst, nil
}

func (id *ULID) UnmarshalBinary(data []byte) error {
	if len(data) != len(*id) {
		return errors.New("invalid data size when unmarshaling").WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}

	copy((*id)[:], data)
	return nil
}

func (id ULID) Bytes() ([]byte, error) {
	return id.MarshalBinary()
}

func (id ULID) UUID() string {
	byteSlice := make([]byte, uuidEncodedSize)

	hex.Encode(byteSlice[0:8], id[0:4])
	byteSlice[8] = '-'
	hex.Encode(byteSlice[9:13], id[4:6])
	byteSlice[13] = '-'
	hex.Encode(byteSlice[14:18], id[6:8])
	byteSlice[18] = '-'
	hex.Encode(byteSlice[19:23], id[8:10])
	byteSlice[23] = '-'
	hex.Encode(byteSlice[24:], id[10:])

	return string(byteSlice)
}

func (id ULID) String() string {
	return id.EncodeString()
}

func (id ULID) EncodeString() string {
	raw, err := id.MarshalText()
	if err != nil {
		return ""
	}

	return *(*string)(unsafe.Pointer(&raw))
}

func (id ULID) EncodeUUID() string {
	buf := make([]byte, uuidEncodedSize-4)
	hex.Encode(buf, id[:])

	return *(*string)(unsafe.Pointer(&buf))
}

func (id *ULID) Scan(src interface{}) (err error) {
	createFormatError := func() error {
		return errors.New("invalid storage format: size must either be 16 bytes or a UUID string").
			WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}

	switch v := src.(type) {
	case []byte:
		switch len(v) {
		case ulid16Bytes:
			return id.UnmarshalBinary(v)
		case ulidUUIDStringLength:
			*id, err = UnmarshalUUID(string(v))
			return err
		default:
			return createFormatError()
		}
	case string:
		if len(v) != ulidUUIDStringLength {
			return createFormatError()
		}
		*id, err = UnmarshalUUID(v)
		return err
	default:
		return createFormatError()
	}
}
func (id ULID) Value() (driver.Value, error) {
	return id.MarshalBinary()
}

func (id ULID) MarshalText() ([]byte, error) {
	dst := make([]byte, textEncodedSize)

	// timestamp
	dst[0] = encoding[(id[0]&224)>>5]
	dst[1] = encoding[id[0]&31]
	dst[2] = encoding[(id[1]&248)>>3]
	dst[3] = encoding[((id[1]&7)<<2)|((id[2]&192)>>6)]
	dst[4] = encoding[(id[2]&62)>>1]
	dst[5] = encoding[((id[2]&1)<<4)|((id[3]&240)>>4)]
	dst[6] = encoding[((id[3]&15)<<1)|((id[4]&128)>>7)]
	dst[7] = encoding[(id[4]&124)>>2]
	dst[8] = encoding[((id[4]&3)<<3)|((id[5]&224)>>5)]
	dst[9] = encoding[id[5]&31]

	// entropy
	dst[10] = encoding[(id[6]&248)>>3]
	dst[11] = encoding[((id[6]&7)<<2)|((id[7]&192)>>6)]
	dst[12] = encoding[(id[7]&62)>>1]
	dst[13] = encoding[((id[7]&1)<<4)|((id[8]&240)>>4)]
	dst[14] = encoding[((id[8]&15)<<1)|((id[9]&128)>>7)]
	dst[15] = encoding[(id[9]&124)>>2]
	dst[16] = encoding[((id[9]&3)<<3)|((id[10]&224)>>5)]
	dst[17] = encoding[id[10]&31]
	dst[18] = encoding[(id[11]&248)>>3]
	dst[19] = encoding[((id[11]&7)<<2)|((id[12]&192)>>6)]
	dst[20] = encoding[(id[12]&62)>>1]
	dst[21] = encoding[((id[12]&1)<<4)|((id[13]&240)>>4)]
	dst[22] = encoding[((id[13]&15)<<1)|((id[14]&128)>>7)]
	dst[23] = encoding[(id[14]&124)>>2]
	dst[24] = encoding[((id[14]&3)<<3)|((id[15]&224)>>5)]
	dst[25] = encoding[id[15]&31]

	return dst, nil
}

func (id *ULID) UnmarshalText(v []byte) error {
	if len(v) == uuidEncodedSize {
		var err error
		*id, err = UnmarshalUUID(string(v))
		return err
	}

	if len(v) != textEncodedSize {
		return errors.New("invalid data size").WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}

	if c2b32[v[0]] == 0xFF ||
		c2b32[v[1]] == 0xFF ||
		c2b32[v[2]] == 0xFF ||
		c2b32[v[3]] == 0xFF ||
		c2b32[v[4]] == 0xFF ||
		c2b32[v[5]] == 0xFF ||
		c2b32[v[6]] == 0xFF ||
		c2b32[v[7]] == 0xFF ||
		c2b32[v[8]] == 0xFF ||
		c2b32[v[9]] == 0xFF ||
		c2b32[v[10]] == 0xFF ||
		c2b32[v[11]] == 0xFF ||
		c2b32[v[12]] == 0xFF ||
		c2b32[v[13]] == 0xFF ||
		c2b32[v[14]] == 0xFF ||
		c2b32[v[15]] == 0xFF ||
		c2b32[v[16]] == 0xFF ||
		c2b32[v[17]] == 0xFF ||
		c2b32[v[18]] == 0xFF ||
		c2b32[v[19]] == 0xFF ||
		c2b32[v[20]] == 0xFF ||
		c2b32[v[21]] == 0xFF ||
		c2b32[v[22]] == 0xFF ||
		c2b32[v[23]] == 0xFF ||
		c2b32[v[24]] == 0xFF ||
		c2b32[v[25]] == 0xFF {
		return errors.New("invalid characters").WithErrorCode(InvalidCharsULIDSystemErrorCode)
	}

	if v[0] > '7' {
		return errors.New("overflow '%b' > 7", v[0]).WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}

	// timestamp (48 bits)
	(*id)[0] = (c2b32[v[0]] << 5) | c2b32[v[1]]
	(*id)[1] = (c2b32[v[2]] << 3) | (c2b32[v[3]] >> 2)
	(*id)[2] = (c2b32[v[3]] << 6) | (c2b32[v[4]] << 1) | (c2b32[v[5]] >> 4)
	(*id)[3] = (c2b32[v[5]] << 4) | (c2b32[v[6]] >> 1)
	(*id)[4] = (c2b32[v[6]] << 7) | (c2b32[v[7]] << 2) | (c2b32[v[8]] >> 3)
	(*id)[5] = (c2b32[v[8]] << 5) | c2b32[v[9]]

	// entropy (80 bits)
	(*id)[6] = (c2b32[v[10]] << 3) | (c2b32[v[11]] >> 2)
	(*id)[7] = (c2b32[v[11]] << 6) | (c2b32[v[12]] << 1) | (c2b32[v[13]] >> 4)
	(*id)[8] = (c2b32[v[13]] << 4) | (c2b32[v[14]] >> 1)
	(*id)[9] = (c2b32[v[14]] << 7) | (c2b32[v[15]] << 2) | (c2b32[v[16]] >> 3)
	(*id)[10] = (c2b32[v[16]] << 5) | c2b32[v[17]]
	(*id)[11] = (c2b32[v[18]] << 3) | c2b32[v[19]]>>2
	(*id)[12] = (c2b32[v[19]] << 6) | (c2b32[v[20]] << 1) | (c2b32[v[21]] >> 4)
	(*id)[13] = (c2b32[v[21]] << 4) | (c2b32[v[22]] >> 1)
	(*id)[14] = (c2b32[v[22]] << 7) | (c2b32[v[23]] << 2) | (c2b32[v[24]] >> 3)
	(*id)[15] = (c2b32[v[24]] << 5) | c2b32[v[25]]

	return nil
}

func (id ULID) time() uint64 {
	return uint64(id[5]) | uint64(id[4])<<8 |
		uint64(id[3])<<16 | uint64(id[2])<<24 |
		uint64(id[1])<<32 | uint64(id[0])<<40
}

func (id *ULID) setTime(t time.Time) error {
	ms := uint64(t.Unix())*1000 +
		uint64(t.Nanosecond()/int(time.Millisecond))

	if ms > maxTime {
		return errors.New("time overflow").WithErrorCode(InvalidTimeFormatULIDSystemErrorCode)
	}

	(*id)[0] = byte(ms >> 40)
	(*id)[1] = byte(ms >> 32)
	(*id)[2] = byte(ms >> 24)
	(*id)[3] = byte(ms >> 16)
	(*id)[4] = byte(ms >> 8)
	(*id)[5] = byte(ms)

	return nil
}

func (id ULID) MarshallUint64() (uint64, error) {
	var res uint64
	raw, err := id.MarshalBinary()
	if err != nil {
		return 0, err
	}

	if len(raw) != 16 {
		return 0, errors.New("invalid raw len(%d)", len(raw)).WithErrorCode(InvalidSizeULIDSystemErrorCode)
	}

	buf := bytes.NewReader(raw[8:16])
	if err := binary.Read(buf, binary.BigEndian, &res); err != nil {
		return res, err
	}

	return res, nil
}
