# nomad-rlnc

A dependency-free Go implementation of random linear network coding over GF(2^8), used by the Nomad experiments.

## Implemented

- fixed-size source symbols,
- systematic and random coded symbols,
- in-network linear re-encoding,
- Gaussian-elimination decoding,
- rank-deficiency detection,
- rejection of useless zero-span re-encoding.

The implementation is intentionally small and readable. It is not tuned for high-throughput networking and has not received independent review.

## Security boundary

RLNC is **not encryption or authentication**. A malicious peer can inject polluted coded symbols unless a higher layer authenticates the coded data/object. Generation identifiers, replay handling, pollution-resistant coding and wire serialization are outside this package.

```bash
go test -race ./...
go vet ./...
```
