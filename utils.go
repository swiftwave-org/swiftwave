package main

import (
	"strings"

	"github.com/google/uuid"
)

func generateLongRandomString(length int) (string, error) {
	numUUIDs := (length + 32) / 33 // Number of UUIDs needed to achieve desired length
	randomString := ""

	for i := 0; i < numUUIDs; i++ {
		uuidObj, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}

		randomString += strings.Replace(uuidObj.String(), "-", "", -1)
	}

	return randomString[:length], nil
}
