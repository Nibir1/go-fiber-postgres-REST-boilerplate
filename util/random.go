package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

var seededRand *rand.Rand

func init() {
	// Seed the local random number generator with the current time
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomInt generates a random integer between a specified minimum and maximum value (inclusive)
func RandomInt(min, max int64) int64 {
	return min + seededRand.Int63n(max-min+1)
}

// RandomString generates a random string of a specified length
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[seededRand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner generates a random owner name (assuming a simple 6-character string)
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money between 0 and 1000 (inclusive)
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates a random currency from a pre-defined list
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[seededRand.Intn(n)]
}

// RandomEmail generates a random email address with a 6-character username and gmail.com domain
func RandomEmail() string {
	return fmt.Sprintf("%s@gmail.com", RandomString(6))
}
