package main

import (
	slogger "github.com/khiemdoan/go-simple-slogger"
)

func main() {
	logger := slogger.NewSlogger()
	logger.Info("Hello, world!")
}
