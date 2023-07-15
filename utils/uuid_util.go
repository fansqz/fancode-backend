package utils

import (
	"github.com/google/uuid"
	"log"
)

func GetUUID() string {
	u1, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	return u1.String()
}
