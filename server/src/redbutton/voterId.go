package redbutton
import (
	"crypto/sha256"
	"math/rand"
	"strconv"
	"fmt"
)


// generate a random voter ID
func uniqueId() string {
	h := sha256.New()
	result := h.Sum([]byte(strconv.Itoa(rand.Int())))

	return fmt.Sprintf("%x", result)
}
