package ulid

import (
	"github.com/google/uuid"
	"github.com/matoous/go-nanoid/v2"
	oulid "github.com/oklog/ulid"
	"testing"
	"time"
)

// Benchmark for NanoID generation
func BenchmarkNanoIDGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gonanoid.Must()
	}
}

// Benchmark for NanoID generation
func BenchmarkAnotherULIDGeneration(b *testing.B) {
	entropy := oulid.Monotonic(defaultEntropy, 2)
	for i := 0; i < b.N; i++ {
		s := oulid.MustNew(oulid.Timestamp(time.Now()), entropy).String()
		_, err := oulid.Parse(s)
		if err != nil {
			panic(err)
		}
	}
}

// Benchmark for ULID generation
func BenchmarkULIDGeneration(b *testing.B) {
	entropy := oulid.Monotonic(defaultEntropy, 2)
	for i := 0; i < b.N; i++ {
		s := MustNew(entropy).String()

		_, err := UnmarshalString(s)
		if err != nil {
			panic(err)
		}
	}
}

// Benchmark for UUID generation
func BenchmarkUUIDGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := uuid.New()
		_ = s.String()
	}
}