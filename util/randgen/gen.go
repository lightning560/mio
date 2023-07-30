package randgen

import (
	// "time"
	"math/rand"
	"miopkg/log"
	"strconv"
	// "math/big"
)

// const (
// 	NUMBER = iota
// 	CHAR
// 	mix
// 	advance
// )

func RandomString(length int, kind string) string {
	passwd := make([]rune, length)
	var codeModel []rune
	switch kind {
	case "num":
		codeModel = []rune("0123456789")
	case "char":
		codeModel = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "mix":
		codeModel = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "advance":
		codeModel = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+=-!@#$%*,.[]")
	default:
		codeModel = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	}
	for i := range passwd {
		index := rand.Int63n(int64(len(codeModel)))
		passwd[i] = codeModel[index]
	}
	return string(passwd)
}

func RandomInt64(length int) int64 {
	str := RandomString(length, "num")
	res, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Errorf("CreateUser uid strconv.ParseInt error:%w", err)
		return -1
	}
	return res
}
