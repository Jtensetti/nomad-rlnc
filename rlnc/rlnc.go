package rlnc

import (
	"crypto/rand"
	"errors"
	"fmt"
)

type Symbol struct {
	Coeff []byte
	Data  []byte
}

type Encoder struct {
	source     [][]byte
	original   int
	symbolSize int
}

func NewEncoder(data []byte, symbolSize int) (*Encoder, error) {
	if symbolSize <= 0 {
		return nil, errors.New("symbolSize must be positive")
	}
	if len(data) == 0 {
		return nil, errors.New("data must not be empty")
	}
	k := (len(data) + symbolSize - 1) / symbolSize
	src := make([][]byte, k)
	for i := 0; i < k; i++ {
		src[i] = make([]byte, symbolSize)
		start := i * symbolSize
		end := start + symbolSize
		if end > len(data) {
			end = len(data)
		}
		copy(src[i], data[start:end])
	}
	return &Encoder{source: src, original: len(data), symbolSize: symbolSize}, nil
}

func (e *Encoder) K() int            { return len(e.source) }
func (e *Encoder) OriginalSize() int { return e.original }
func (e *Encoder) SymbolSize() int   { return e.symbolSize }

func (e *Encoder) Encode() (Symbol, error) {
	coeff := make([]byte, e.K())
	for {
		if _, err := rand.Read(coeff); err != nil {
			return Symbol{}, err
		}
		nonzero := false
		for _, c := range coeff {
			if c != 0 {
				nonzero = true
				break
			}
		}
		if nonzero {
			break
		}
	}
	return e.encodeWith(coeff), nil
}

func (e *Encoder) Systematic(i int) (Symbol, error) {
	if i < 0 || i >= e.K() {
		return Symbol{}, fmt.Errorf("index out of range")
	}
	coeff := make([]byte, e.K())
	coeff[i] = 1
	return e.encodeWith(coeff), nil
}

func (e *Encoder) encodeWith(coeff []byte) Symbol {
	out := make([]byte, e.symbolSize)
	for i, c := range coeff {
		if c == 0 {
			continue
		}
		for j := range out {
			out[j] ^= mul(c, e.source[i][j])
		}
	}
	return Symbol{Coeff: append([]byte(nil), coeff...), Data: out}
}

func ReEncode(symbols []Symbol) (Symbol, error) {
	if len(symbols) == 0 {
		return Symbol{}, errors.New("no symbols")
	}
	k := len(symbols[0].Coeff)
	size := len(symbols[0].Data)
	for _, s := range symbols {
		if len(s.Coeff) != k || len(s.Data) != size {
			return Symbol{}, errors.New("incompatible symbols")
		}
	}
	weights := make([]byte, len(symbols))
	for {
		if _, err := rand.Read(weights); err != nil {
			return Symbol{}, err
		}
		nz := false
		for _, w := range weights {
			if w != 0 {
				nz = true
				break
			}
		}
		if nz {
			break
		}
	}
	coeff := make([]byte, k)
	data := make([]byte, size)
	for i, s := range symbols {
		w := weights[i]
		if w == 0 {
			continue
		}
		for j := range coeff {
			coeff[j] ^= mul(w, s.Coeff[j])
		}
		for j := range data {
			data[j] ^= mul(w, s.Data[j])
		}
	}
	return Symbol{Coeff: coeff, Data: data}, nil
}

func Decode(symbols []Symbol, k, originalSize int) ([]byte, error) {
	if k <= 0 || originalSize <= 0 {
		return nil, errors.New("invalid dimensions")
	}
	if len(symbols) < k {
		return nil, errors.New("insufficient symbols")
	}
	size := len(symbols[0].Data)
	rows := make([][]byte, 0, len(symbols))
	for _, s := range symbols {
		if len(s.Coeff) != k || len(s.Data) != size {
			return nil, errors.New("incompatible symbols")
		}
		row := make([]byte, k+size)
		copy(row, s.Coeff)
		copy(row[k:], s.Data)
		rows = append(rows, row)
	}

	rank := 0
	pivots := make([]int, 0, k)
	for col := 0; col < k && rank < len(rows); col++ {
		pivot := -1
		for r := rank; r < len(rows); r++ {
			if rows[r][col] != 0 {
				pivot = r
				break
			}
		}
		if pivot == -1 {
			continue
		}
		rows[rank], rows[pivot] = rows[pivot], rows[rank]
		scale := inv(rows[rank][col])
		for c := col; c < len(rows[rank]); c++ {
			rows[rank][c] = mul(rows[rank][c], scale)
		}
		for r := 0; r < len(rows); r++ {
			if r == rank || rows[r][col] == 0 {
				continue
			}
			factor := rows[r][col]
			for c := col; c < len(rows[r]); c++ {
				rows[r][c] = add(rows[r][c], mul(factor, rows[rank][c]))
			}
		}
		pivots = append(pivots, col)
		rank++
		if rank == k {
			break
		}
	}
	if rank < k {
		return nil, errors.New("rank deficient symbol set")
	}

	decoded := make([][]byte, k)
	for r, col := range pivots[:k] {
		decoded[col] = append([]byte(nil), rows[r][k:]...)
	}
	out := make([]byte, 0, k*size)
	for i := 0; i < k; i++ {
		if decoded[i] == nil {
			return nil, errors.New("missing pivot")
		}
		out = append(out, decoded[i]...)
	}
	if originalSize > len(out) {
		return nil, errors.New("original size exceeds decoded buffer")
	}
	return out[:originalSize], nil
}
