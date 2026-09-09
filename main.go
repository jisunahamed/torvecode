package main

import (
	"github.com/jisunahamed/torvecode/cmd"
	"github.com/jisunahamed/torvecode/internal/logging"
)

func main() {
	defer logging.RecoverPanic("main", func() {
		logging.ErrorPersist("Application terminated due to unhandled panic")
	})

	cmd.Execute()
}
