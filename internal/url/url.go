package url

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func ShortenURL(originalUrl string) string {
	hash := sha256.Sum256([]byte(originalUrl))
	return base64.URLEncoding.EncodeToString(hash[:][:7])
}

func NormalizeURL(url string) string {
	if strings.HasPrefix(url, "www.") {
		return url[len("www."):]
	}

	if strings.HasPrefix(url, "http://www.") {
		return url[len("http://www."):]
	}
	if strings.HasPrefix(url, "http://") {
		return url[len("http://"):]
	}

	if strings.HasPrefix(url, "https://www.") {
		return url[len("https://www."):]
	}
	if strings.HasPrefix(url, "https://") {
		return url[len("https://"):]
	}

	return url
}
