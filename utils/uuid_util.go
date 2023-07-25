package utils

import (
	"github.com/google/uuid"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func GetUUID() string {
	u1, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	return u1.String()
}

// 通过时间搓 + 随机数生成的较短的随机code
func GetGenerateUniqueCode() string {
	timestamp := time.Now().Unix()
	randomNum := rand.Intn(1000) // 生成一个0到999之间的随机数

	uniqueNumber := strconv.FormatInt(timestamp, 10) + strconv.Itoa(randomNum)
	return uniqueNumber
}
