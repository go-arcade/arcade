package tool

import "encoding/base64"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 15:57
 * @file: base64.go
 * @description: base64
 */

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
