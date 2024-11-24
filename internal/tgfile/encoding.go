package tgfile

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"unsafe"
)

func StrToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func B64Str(s string) string {
	return base64.StdEncoding.EncodeToString(StrToBytes(s))
}

func HashBytes(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
