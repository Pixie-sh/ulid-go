## ULID

- ULID compatible
- UUIDv4 compatible
- Time based
- with entropy abstraction optimized

### Supported Database Drivers
- pgx (coming soon)
- gorm (coming soon)
- sql.DB (coming soon)
             
### Benchmark
```
goos: darwin
goarch: arm64
pkg: github.com/pixie-sh/ulid-go
cpu: Apple M3 Max
BenchmarkNanoIDGeneration
BenchmarkNanoIDGeneration-16         	 3062644	       388.9 ns/op
BenchmarkAnotherULIDGeneration
BenchmarkAnotherULIDGeneration-16    	 4756365	       253.4 ns/op
BenchmarkULIDGeneration
BenchmarkULIDGeneration-16           	 6347804	       176.8 ns/op
BenchmarkUUIDGeneration
BenchmarkUUIDGeneration-16           	 3714291	       325.3 ns/op
PASS
```

### Thank you
- github.com/google/uuid
- github.com/matoous/go-nanoid/v2
- github.com/oklog/ulid
- github.com/RobThree/NUlid
- github.com/segmentio/ksuid