package utils

import (
	"encoding/base64"
)

func GenerateShortID(uuidString string) string {
	uuidBytes := []byte(uuidString)

	// Encode bytes to base64
	base64Encoded := base64.URLEncoding.EncodeToString(uuidBytes)

	return base64Encoded[:12]
}
