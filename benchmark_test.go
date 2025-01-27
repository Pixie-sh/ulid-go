package pulid

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
		_ = gonanoid.Must(16)
	}
}

// Benchmark for NanoID generation
func BenchmarkAnotherULIDGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := oulid.MustNew(oulid.Timestamp(time.Now()), defaultEntropy).String()
		_, err := oulid.Parse(s)
		if err != nil {
			panic(err)
		}
	}
}

// Benchmark for ULID generation
func BenchmarkPULIDGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := MustNew().String()

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