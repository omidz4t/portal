package main

import (
	"fmt"
	"os"

	"github.com/omidz4t/portal/internal/bot"
)

// version is set at release build time via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-V", "version":
			fmt.Println(version)
			return
		}
	}
	// Expose version to bot package via env for any runtime banners.
	_ = os.Setenv("TGPORTAL_VERSION", version)
	if err := bot.Run(); err != nil {
		os.Exit(1)
	}
}
