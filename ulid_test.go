package pulid

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

func TestULIDGeneration(t *testing.T) {
	id1, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	id2, err := NewScoped(567)
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	if bytes.Equal(id1[:], id2[:]) {
		t.Fatalf("ULIDs are not unique: %v and %v", id1, id2)
	}

	if len(id1[:]) != 16 {
		t.Fatalf("ULID length is incorrect, expected 16 bytes, got %d", len(id1[:]))
	}

	if scope, err := id1.Scope(); err != nil || scope != MaxScopeValue {
		t.Fatalf("ULID scope is incorrect, expected %d got %d; err %+v", MaxScopeValue, scope, err)
	}

	if scope, err := id2.Scope(); err != nil || scope != 567 {
		t.Fatalf("ULID scope is incorrect, expected %d got %d; err %+v", MaxScopeValue, scope, err)
	}

	t.Logf("Generated ULID: %s", id1.String())
	t.Logf("Generated ULID: %s", id2.String())
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

	t.Logf("ULID to UUID conversion successful: %s - %s", id.String(), uuidStr)
}

func TestULIDFormatComparison(t *testing.T) {
	entropy := ulid.Monotonic(rand.Reader, 0)
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

	ulidTime := time.Unix(int64(id.Epoch()/1000), int64(id.Epoch()%1000)*int64(time.Millisecond))
	if !ulidTime.Truncate(time.Millisecond).Equal(now.Truncate(time.Millisecond)) {
		t.Fatalf("Timestamp mismatch: expected %v, got %v", now, ulidTime)
	}

	t.Logf("ULID timestamp validation successful: %v", ulidTime)
}

func TestULIDToUUIDUniqueness(t *testing.T) {
	id1, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}
	uuid1 := id1.UUID()

	id2, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}
	uuid2 := id2.UUID()

	if uuid1 == uuid2 {
		t.Fatalf("UUIDs derived from ULIDs are not unique: %s and %s", uuid1, uuid2)
	}

	t.Logf("Generated unique UUIDs: %s and %s", uuid1, uuid2)
}
func TestUUIDTrailingZeroIssue(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}

	uuidStr := id.UUID()

	if len(uuidStr) != 36 {
		t.Fatalf("UUID length mismatch: expected 36, got %d", len(uuidStr))
	}

	if uuidStr[len(uuidStr)-1] == '0' && uuidStr[len(uuidStr)-2] == '0' {
		t.Fatalf("UUID ends with unexpected trailing zeros: %s", uuidStr)
	}

	t.Logf("UUID generated successfully: %s", uuidStr)
}

func TestUUIDToULIDConversion(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}
	uuidStr := id.UUID()

	convertedULID, err := UnmarshalUUID(uuidStr)
	if err != nil {
		t.Fatalf("Failed to convert UUID back to ULID: %v", err)
	}

	if id != convertedULID {
		t.Fatalf("ULID mismatch after UUID conversion: expected %s, got %s", id, convertedULID)
	}

	t.Logf("UUID to ULID conversion successful: UUID - %s, ULID - %s", uuidStr, convertedULID.String())
}

func TestULIDToUUIDCompliance(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("Failed to generate ULID: %v", err)
	}
	uuidStr := id.UUID()

	if len(uuidStr) != 36 {
		t.Fatalf("Generated UUID length is incorrect: expected 36, got %d", len(uuidStr))
	}

	uuidRegex := `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
	if !regexp.MustCompile(uuidRegex).MatchString(uuidStr) {
		t.Fatalf("Generated UUID does not comply with UUIDv4 format: %s", uuidStr)
	}

	t.Logf("ULID to UUID compliance successful: %s", uuidStr)
}

func TestUUIDUniquenessAfterConversion(t *testing.T) {
	ulidCount := 1000
	uuidSet := make(map[string]bool)

	for i := 0; i < ulidCount; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("Failed to generate ULID: %v", err)
		}
		uuidStr := id.UUID()

		if uuidSet[uuidStr] {
			t.Fatalf("Duplicate UUID detected: %s", uuidStr)
		}
		uuidSet[uuidStr] = true
	}

	t.Logf("All UUIDs generated from ULIDs are unique for %d iterations; %+v", ulidCount, uuidSet)
}

func TestULIDToUUIDMassiveUniqueness(t *testing.T) {
	const totalIDs = 10_000_000 // 10 million ULIDs
	uuidSet := make(map[string]struct{}, totalIDs)

	for i := 0; i < totalIDs; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("Failed to generate ULID: %v", err)
		}

		uuidStr := id.UUID()
		if _, exists := uuidSet[uuidStr]; exists {
			t.Fatalf("Duplicate UUID detected after %d iterations: %s", i, uuidStr)
		}

		uuidSet[uuidStr] = struct{}{}

		// Log progress every 1 million iterations
		if (i+1)%1_000_000 == 0 {
			t.Logf("Generated %d unique UUIDs so far", i+1)
		}
	}

	t.Logf("Successfully generated %d unique UUIDs", totalIDs)
}

func TestConcurrentULIDToUUIDUniqueness(t *testing.T) {
	const totalIDs = 10_000_000 // Total number of ULIDs to generate
	const numWorkers = 20       // Number of concurrent workers
	idsPerWorker := totalIDs / numWorkers

	// Use a thread-safe map for uniqueness checks
	uuidSet := sync.Map{}
	ulidSet := sync.Map{}

	// Define a worker function
	worker := func(start, count int, wg *sync.WaitGroup, errChan chan error) {
		defer wg.Done()

		for i := 0; i < count; i++ {
			id, err := New()
			if err != nil {
				errChan <- fmt.Errorf("worker %d failed to generate ULID: %v", start, err)
				return
			}

			uuidStr := id.UUID()

			// Check uniqueness in a thread-safe manner
			entry := fmt.Sprintf("worker-%d;Epoch:%s", start, time.Now().UTC().String())
			if existing, exists := ulidSet.LoadOrStore(id, entry); exists {
				errChan <- fmt.Errorf("duplicate ULID detected: %s worker %s <-> existing: %s", id.String(), entry, existing.(string))
				return
			}

			if existing, exists := uuidSet.LoadOrStore(uuidStr, entry); exists {
				errChan <- fmt.Errorf("duplicate UUID detected: %s worker %s <-> existing: %s", uuidStr, entry, existing.(string))
				return
			}
		}
	}

	// Create a wait group for worker synchronization
	var wg sync.WaitGroup
	errChan := make(chan error, numWorkers)

	// Launch workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i*idsPerWorker, idsPerWorker, &wg, errChan)
	}

	// Wait for all workers to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	t.Logf("Successfully generated %d unique UUIDs across %d workers", totalIDs, numWorkers)
}

func TestULIDParsing(t *testing.T) {
	// Test parsing a valid ULID string
	original, _ := New()
	str := original.String()

	parsed, err := UnmarshalString(str)
	if err != nil {
		t.Fatalf("Failed to parse valid ULID string: %v", err)
	}

	if parsed != original {
		t.Fatalf("Parsed ULID doesn't match original: %v != %v", parsed, original)
	}

	// Test parsing an invalid string
	_, err = UnmarshalString("invalid-ulid-string")
	if err == nil {
		t.Fatalf("Expected error when parsing invalid ULID string")
	}
}

func TestULIDMustFunctions(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustParse with invalid input should panic")
		}
	}()

	// Test successful MustParse
	original := MustNew()
	str := original.String()
	parsed, _ := UnmarshalString(str)
	if parsed != original {
		t.Fatalf("MustParse result doesn't match original")
	}
	 _, err := UnmarshalString("invalid-ulid-string")
	if err != nil {
		panic(err)
	}
}

func TestULIDScopeEdgeCases(t *testing.T) {
	// Test with scope of 0
	id, err := NewScoped(ZeroedScopeValue)
	if err != nil {
		t.Fatalf("Failed to create ULID with scope 0: %v", err)
	}
	scope, _ := id.Scope()
	if scope != MaxScopeValue {
		t.Fatalf("Expected scope MaxScopeValue, got %d", scope)
	}

	// Test with max valid scope
	id, err = NewScoped(MaxScopeValue)
	if err != nil {
		t.Fatalf("Failed to create ULID with max scope: %v", err)
	}
	scope, _ = id.Scope()
	if scope != MaxScopeValue {
		t.Fatalf("Expected scope %d, got %d", MaxScopeValue, scope)
	}

	// Test with invalid scope (too large)
	u, err := NewScoped(MaxScopeValue + 1)
	if err != nil {
		t.Fatalf("no error expected as compiler trim its to zero")
	}

	if s, _ := u.Scope(); s != MaxScopeValue {
		t.Fatalf("Expected scope %d, got %d", MaxScopeValue, s)
	}
}

func TestULIDConcurrency(t *testing.T) {
	const goroutines = 100
	const ulidsPerRoutine = 100

	var wg sync.WaitGroup
	idChan := make(chan ULID, goroutines*ulidsPerRoutine)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < ulidsPerRoutine; j++ {
				id, err := New()
				if err != nil {
					t.Errorf("ULID generation failed: %v", err)
					return
				}
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Check for uniqueness
	seen := make(map[ULID]bool)
	for id := range idChan {
		if seen[id] {
			t.Fatalf("Duplicate ULID detected: %v", id)
		}
		seen[id] = true
	}
}

func TestULIDJSON(t *testing.T) {
	id, _ := New()

	// Test marshalling
	jsonData, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("JSON marshalling failed: %v", err)
	}

	// Test unmarshalling
	var unmarshalled ULID
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("JSON unmarshalling failed: %v", err)
	}

	if id != unmarshalled {
		t.Fatalf("JSON round trip failed: expected %v, got %v", id, unmarshalled)
	}

	// Test unmarshalling from a string
	jsonString := fmt.Sprintf(`"%s"`, id.String())
	if err := json.Unmarshal([]byte(jsonString), &unmarshalled); err != nil {
		t.Fatalf("JSON unmarshalling from string failed: %v", err)
	}

	if id != unmarshalled {
		t.Fatalf("JSON string unmarshalling failed: expected %v, got %v", id, unmarshalled)
	}
}

func TestULIDTimestampEdgeCases(t *testing.T) {
	uFromUuid, _ := UnmarshalUUID("019672bc-73c3-ffff-9e66-1b0983a0ec48")

	fmt.Println(MustNew())
	fmt.Println(uFromUuid)
	fmt.Printf("%v",MustNew())
	fmt.Printf("\n %u",MustNew())
	fmt.Printf("\n %v",uFromUuid)

	n := MustNew()
	if fmt.Sprintf("%u", n) != n.UUID() {
		t.Fatalf("ULID u fromatter and UUID should be equal")
	}
}