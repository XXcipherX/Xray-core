package splithttp

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
)

const (
	SeqEncodingDecimal   = "decimal"
	SeqEncodingMaskedHex = "masked-hex"
)

const seqFeistelRounds = 6

func (c *Config) GetNormalizedSeqEncoding() string {
	if c == nil || c.SeqEncoding == "" {
		return SeqEncodingDecimal
	}
	return c.SeqEncoding
}

func (c *Config) EncodeSeq(sessionID string, seq uint64) string {
	switch c.GetNormalizedSeqEncoding() {
	case SeqEncodingMaskedHex:
		return encodeMaskedHexSeq(seq, sessionID)
	default:
		return strconv.FormatUint(seq, 10)
	}
}

func (c *Config) DecodeSeq(sessionID string, seqStr string) (uint64, error) {
	switch c.GetNormalizedSeqEncoding() {
	case SeqEncodingMaskedHex:
		if len(seqStr) != 16 {
			return 0, fmt.Errorf("invalid masked-hex seq length: %d", len(seqStr))
		}
		var encoded [8]byte
		_, err := hex.Decode(encoded[:], []byte(seqStr))
		if err != nil {
			return 0, err
		}
		seq := binary.BigEndian.Uint64(encoded[:])
		return unpermuteSeq(seq, sessionID), nil
	default:
		return strconv.ParseUint(seqStr, 10, 64)
	}
}

func encodeMaskedHexSeq(seq uint64, sessionID string) string {
	var encoded [16]byte
	var permuted [8]byte
	binary.BigEndian.PutUint64(permuted[:], permuteSeq(seq, sessionID))
	hex.Encode(encoded[:], permuted[:])
	return string(encoded[:])
}

func seqFeistelKeys(sessionID string) [seqFeistelRounds]uint32 {
	sum := sha256.Sum256([]byte("xhttp packet-up seq:" + sessionID))
	var keys [seqFeistelRounds]uint32
	for i := range keys {
		keys[i] = binary.LittleEndian.Uint32(sum[i*4:])
	}
	return keys
}

func seqRound(v uint32, key uint32) uint32 {
	v += key
	v ^= v >> 16
	v *= 0x7feb352d
	v ^= v >> 15
	v *= 0x846ca68b
	v ^= v >> 16
	return v
}

func permuteSeq(seq uint64, sessionID string) uint64 {
	left := uint32(seq >> 32)
	right := uint32(seq)

	for _, key := range seqFeistelKeys(sessionID) {
		left, right = right, left^seqRound(right, key)
	}

	return uint64(left)<<32 | uint64(right)
}

func unpermuteSeq(seq uint64, sessionID string) uint64 {
	left := uint32(seq >> 32)
	right := uint32(seq)
	keys := seqFeistelKeys(sessionID)

	for i := len(keys) - 1; i >= 0; i-- {
		left, right = right^seqRound(left, keys[i]), left
	}

	return uint64(left)<<32 | uint64(right)
}
