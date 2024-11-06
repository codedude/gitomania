package main

import (
	"gitomania/internal/file"
	"log/slog"
)

func main() {
	var _ = file.IsTrue(true)
	slog.Info("Here")
}
