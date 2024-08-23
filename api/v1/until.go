package v1

import (
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

func GenerateNumber(length int) string {

	const chars = "QWERTYUIOPASDFGHJKLZXCVBNM0123456789"
	var result strings.Builder
	result.Grow(length)

	// 使用当前时间作为种子
	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 0; i < length; i++ {
		result.WriteByte(chars[rand.Intn(len(chars))])
	}

	return result.String()

}
