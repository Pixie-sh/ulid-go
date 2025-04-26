module github.com/pixie-sh/ulid-go

go 1.23.0

require (
	github.com/google/uuid v1.6.0
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/oklog/ulid v1.3.1
	github.com/pixie-sh/errors-go v0.3.3
)

require (
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pixie-sh/logger-go v0.4.2 // indirect
	golang.org/x/crypto v0.37.0 // indirect
)

//replace github.com/pixie-sh/errors-go => ../errors-go
