package utils

import (
	"log"
	"os"
)

func NewLogger() *log.Logger {
	logger := log.New(
		os.Stdout,
		"",
		log.Ldate|
			log.Ltime|
			log.Lmicroseconds|
			log.LUTC|
			log.Lshortfile,
	)
	return logger
}
