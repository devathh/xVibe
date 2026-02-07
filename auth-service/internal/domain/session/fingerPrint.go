package session

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Fingerprint string

func (fp Fingerprint) Value() string {
	return string(fp)
}

func (fp Fingerprint) Compare(clientIP, userAgent string) bool {
	customFP := NewFingerPrint(clientIP, userAgent)
	return customFP.Value() == fp.Value()
}

func NewFingerPrint(clientIP, userAgent string) Fingerprint {
	hash := sha256.Sum256(fmt.Appendf(nil, "%s:%s", clientIP, userAgent[:min(len(userAgent), 16)]))
	return Fingerprint(hex.EncodeToString(hash[:]))
}
