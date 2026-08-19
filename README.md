# nomad-rlnc

A compact, dependency-free Go implementation of **random linear network coding (RLNC)** over GF(2^8) for Nomad research.

Implemented:

- fixed-size source symbols,
- systematic symbols,
- random coded symbols,
- in-network re-encoding,
- Gaussian-elimination decoding,
- rank-deficiency detection,
- property-style round-trip tests.

This is real executable code, but it has not received independent performance or security review. RLNC itself is not encryption; callers must authenticate and encrypt sensitive content separately.

## Build

```bash
go test ./...
go test -race ./...
go vet ./...
```

## API sketch

```go
enc, _ := rlnc.NewEncoder(data, 1024)
s, _ := enc.Encode()
re, _ := rlnc.ReEncode([]rlnc.Symbol{s, other})
plain, err := rlnc.Decode(symbols, enc.K(), enc.OriginalSize())
```
