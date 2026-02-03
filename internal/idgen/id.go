package idgen

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

const (
	PrefixEvent    = "E"
	PrefixLesson   = "L"
	PrefixDecision = "D"
	PrefixPattern  = "P"
	PrefixCustom   = "K"
)

var (
	v2IDPattern     = regexp.MustCompile(`^[ELDPK]-[0-9A-HJKMNP-TV-Z]{26}$`)
	v2ItemIDPattern = regexp.MustCompile(`^[LDPK]-[0-9A-HJKMNP-TV-Z]{26}$`)
	crockford32     = []byte("0123456789ABCDEFGHJKMNPQRSTVWXYZ")
	base32          = big.NewInt(32)
)

// NewItemID returns a new typed ID for a knowledge item type.
func NewItemID(itemType string) (string, error) {
	prefix, err := PrefixForType(itemType)
	if err != nil {
		return "", err
	}
	return NewTyped(prefix)
}

// PrefixForType maps a dory item type to an ID prefix.
func PrefixForType(itemType string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(itemType)) {
	case "lesson":
		return PrefixLesson, nil
	case "decision":
		return PrefixDecision, nil
	case "pattern":
		return PrefixPattern, nil
	default:
		if itemType == "" {
			return "", errors.New("item type is required")
		}
		return PrefixCustom, nil
	}
}

// NewEventID returns a new event ID.
func NewEventID() (string, error) {
	return NewTyped(PrefixEvent)
}

// NewTyped returns a new typed ULID-style ID (e.g., L-01JX...).
func NewTyped(prefix string) (string, error) {
	return NewTypedAt(prefix, time.Now().UTC())
}

// NewTypedAt returns a new typed ULID-style ID using the provided timestamp.
func NewTypedAt(prefix string, at time.Time) (string, error) {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	if len(prefix) != 1 {
		return "", fmt.Errorf("invalid prefix %q: expected one letter", prefix)
	}
	if !strings.Contains("ELDPK", prefix) {
		return "", fmt.Errorf("invalid prefix %q: expected one of E,L,D,P,K", prefix)
	}

	var raw [16]byte
	ms := uint64(at.UTC().UnixMilli())
	raw[0] = byte(ms >> 40)
	raw[1] = byte(ms >> 32)
	raw[2] = byte(ms >> 24)
	raw[3] = byte(ms >> 16)
	raw[4] = byte(ms >> 8)
	raw[5] = byte(ms)

	if _, err := rand.Read(raw[6:]); err != nil {
		return "", fmt.Errorf("failed generating random ULID entropy: %w", err)
	}

	return prefix + "-" + encodeCrockford32(raw[:]), nil
}

// IsValidV2ID returns true for any v2 typed ID (event or item).
func IsValidV2ID(id string) bool {
	return v2IDPattern.MatchString(strings.TrimSpace(id))
}

// IsValidItemID returns true for v2 typed item IDs.
func IsValidItemID(id string) bool {
	return v2ItemIDPattern.MatchString(strings.TrimSpace(id))
}

func encodeCrockford32(raw []byte) string {
	n := new(big.Int).SetBytes(raw)
	mod := new(big.Int)
	out := make([]byte, 26)

	for i := len(out) - 1; i >= 0; i-- {
		n.DivMod(n, base32, mod)
		out[i] = crockford32[mod.Int64()]
	}

	return string(out)
}
