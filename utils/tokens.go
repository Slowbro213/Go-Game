package utils

import (

	"crypto/rand"
	"encoding/base64"

)


func GenerateToken() string {
	b := make([]byte, 32)
	_ , err  := rand.Read(b)
	if err != nil {
		panic("Couldnt Generate Token")
	}

	return base64.URLEncoding.EncodeToString(b)
}

