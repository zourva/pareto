package box

import "encoding/base64"

// Base64 wraps the golang default base64 encoder.
func Base64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}
