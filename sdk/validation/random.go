package validation

import (
	"crypto/rand"
	"math/big"
)

// SecureRandomInt 生成安全的随机整数
func SecureRandomInt(min, max int) (int, error) {
	if min > max {
		return 0, ErrInvalidArgument("min 必须小于等于 max")
	}
	if min == max {
		return min, nil
	}
	rangeSize := int64(max - min + 1)
	maxBig := big.NewInt(rangeSize)
	n, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + min, nil
}
