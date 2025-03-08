package lib

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateSecureVerificationCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(90000))
	code := n.Int64() + 10000
	return fmt.Sprintf("%05d", code)
}
