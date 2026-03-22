package utils

import (
	"math/rand/v2"
	"os"
	"strconv"
)

func BoolPtr(b bool) *bool {
	return &b
}

func GetEnvInt(key string, fallback int) int {
	if raw, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(raw)
		if err == nil {
			return i
		}
	}
	return fallback
}

func GenerateLobbyId() int64 {
	// 26^8
	const maxLobbies int64 = 208827064576

	// Uint64N is the new standard; it's type-safe and avoids
	// the "modulo bias" that older random implementations had.
	return int64(rand.Uint64N(uint64(maxLobbies)))
}

func IdToString(id int64) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = charset[id%26]
		id /= 26
	}
	return string(b)
}

func StringToId(code string) int64 {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var id int64

	// We process the string from left to right
	for i := 0; i < len(code); i++ {
		// Find the index of the character in our charset
		// (A=0, B=1, etc.)
		charVal := int64(code[i] - 'A')

		// Shift the existing ID up by one power of 26
		// and add the new character's value
		id = id*26 + charVal
	}

	return id
}

func GenLobbyCode(n int64) string {
	var res []byte
	for n > 0 {
		n-- // Adjust for 1-based indexing
		res = append([]byte{byte('A' + (n % 26))}, res...)
		n /= 26
	}
	return string(res)
}
