package utils

import (

	"crypto/rand"
	"encoding/base64"

)


func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

