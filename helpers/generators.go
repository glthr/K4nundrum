package helpers

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// Split splits the ciphertext based on a separator
// and returns non-empty segments
func Split(ciphertext string, separator rune) []string {
	return strings.FieldsFunc(ciphertext, func(s rune) bool {
		return s == separator
	})
}

// GenerateRandomString generates pseudo-K4s
func GenerateRandomString(size int) string {
	charSet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var sb strings.Builder
	sb.Grow(size)

	for i := 0; i < size; i++ {
		randomIndex, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(charSet))),
		)
		if err != nil {
			panic(err)
		}
		sb.WriteByte(charSet[randomIndex.Int64()])
	}

	return sb.String()
}

func permute(input []string, start int, ch chan []string) {
	if start == len(input)-1 {
		temp := make([]string, len(input))
		copy(temp, input)
		ch <- temp
		return
	}

	for i := start; i < len(input); i++ {
		input[start], input[i] = input[i], input[start]
		permute(input, start+1, ch)
		input[start], input[i] = input[i], input[start]
	}
}

// GeneratePermutations generates permutations of text segments
func GeneratePermutations(input []string) <-chan []string {
	ch := make(chan []string)
	go func() {
		defer close(ch)
		permute(input, 0, ch)
	}()
	return ch
}
