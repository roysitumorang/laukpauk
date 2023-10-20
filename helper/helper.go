package helper

import (
	"crypto/rand"
	"math/big"
	"unsafe"

	"github.com/bwmarrin/snowflake"
	"github.com/speps/go-hashids/v2"
)

const (
	letters = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var (
	snowflakeNode *snowflake.Node
)

func GenerateRandomString(length int) string {
	bs := make([]byte, length)
	max := big.NewInt(int64(len(letters)))
	for i := 0; i < length; {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			continue
		}
		bs[i] = letters[num.Int64()]
		i++
	}
	return ByteSlice2String(bs)
}

func String2ByteSlice(str string) []byte {
	if str == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

func ByteSlice2String(bs []byte) string {
	n := len(bs)
	if n == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(bs), n)
}

func GenerateHashIDs(mingLength int, numbers ...int64) (string, error) {
	data := hashids.NewData()
	data.MinLength = mingLength
	hashID, err := hashids.NewWithData(data)
	if err != nil {
		return "", err
	}
	hash, err := hashID.EncodeInt64(numbers)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func GenerateSnowflakeUniqueID() (_ int64, err error) {
	if snowflakeNode == nil {
		if snowflakeNode, err = snowflake.NewNode(1); err != nil {
			return
		}
	}
	return snowflakeNode.Generate().Int64(), nil

}
