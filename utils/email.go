package utils

import "log"

func SendWarningEmail(userID string) {
	log.Printf("ID %s был изменен", userID)
}
