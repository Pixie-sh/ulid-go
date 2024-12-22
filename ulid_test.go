package ulid

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
	"math/rand"
)

func TestULIDGeneration(t *testing.T) {
	id1, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	id2, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	if bytes.Equal(id1[:], id2[:]) {
		t.Fatalf("ULIDs are not unique: %v and %v", id1, id2)
	}

	if len(id1[:]) != 16 {
		t.Fatalf("ULID length is incorrect, expected 16 bytes, got %d", len(id1[:]))
	}

	t.Logf("Generated ULID: %s", id1.String())
}

func TestULIDToUUIDConversion(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	uuidStr := id.UUID()
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		t.Fatalf("Failed to parse ULID as UUID: %v", err)
	}

	if parsedUUID.String() != uuidStr {
		t.Fatalf("UUID string mismatch: expected %s, got %s", uuidStr, parsedUUID.String())
	}

	t.Logf("ULID to UUID conversion successful: %s", uuidStr)
}

func TestULIDFormatComparison(t *testing.T) {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	thirdPartyULID := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	thirdPartyStr := thirdPartyULID.String()

	myULID := MustNew()
	myULIDString := myULID.String()

	ulidRegex := `^[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}$`

	re := regexp.MustCompile(ulidRegex)

	if !re.MatchString(myULIDString) {
		t.Fatalf("Your ULID does not match the ULID format: %s", myULIDString)
	}

	if !re.MatchString(thirdPartyStr) {
		t.Fatalf("Third-party ULID does not match the ULID format: %s", thirdPartyStr)
	}

	t.Logf("Both ULIDs have valid formats: Your ULID - %s, Third-party ULID - %s", myULIDString, thirdPartyStr)
}

func TestULIDMarshalling(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	binaryData, err := id.MarshalBinary()
	if err != nil {
		t.Fatalf("Binary marshalling failed: %v", err)
	}

	var unmarshalledID ULID
	if err := unmarshalledID.UnmarshalBinary(binaryData); err != nil {
		t.Fatalf("Binary unmarshalling failed: %v", err)
	}

	if id != unmarshalledID {
		t.Fatalf("Binary marshalling mismatch: expected %v, got %v", id, unmarshalledID)
	}

	textData, err := id.MarshalText()
	if err != nil {
		t.Fatalf("Text marshalling failed: %v", err)
	}

	if err := unmarshalledID.UnmarshalText(textData); err != nil {
		t.Fatalf("Text unmarshalling failed: %v", err)
	}

	if id != unmarshalledID {
		t.Fatalf("Text marshalling mismatch: expected %v, got %v", id, unmarshalledID)
	}
}

func TestULIDErrorHandling(t *testing.T) {
	var id ULID

	invalidBinary := []byte{0x01, 0x02}
	if err := id.UnmarshalBinary(invalidBinary); err == nil {
		t.Fatalf("Expected error for invalid binary data, got nil")
	}

	invalidText := []byte("this-is-not-a-ulid")
	if err := id.UnmarshalText(invalidText); err == nil {
		t.Fatalf("Expected error for invalid text data, got nil")
	}
}

func TestULIDTimestamp(t *testing.T) {
	now := time.Now()
	id, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	ulidTime := time.Unix(int64(id.time()/1000), int64(id.time()%1000)*int64(time.Millisecond))
	if !ulidTime.Truncate(time.Millisecond).Equal(now.Truncate(time.Millisecond)) {
		t.Fatalf("Timestamp mismatch: expected %v, got %v", now, ulidTime)
	}

	t.Logf("ULID timestamp validation successful: %v", ulidTime)
}