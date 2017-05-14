package main

import (
	"crypto/md5"
	"encoding/hex"
)

func md5hash(rawstr string) string {
	hasher := md5.New()
	hasher.Write([]byte(rawstr))
	return hex.EncodeToString(hasher.Sum(nil))
}
