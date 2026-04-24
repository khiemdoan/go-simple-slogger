package main

import (
	slogger "github.com/khiemdoan/go-simple-slogger"
)

func main() {
	logger := slogger.NewAsyncSlogger()
	logger.Info("Hello, world!")
}
