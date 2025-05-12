package main

import (
	"flag"

	"github.com/brianliurabbithole/gitlog/logger"
)

func main() {
	logger.Init()
	defer logger.GetLogger().Sync()

	var folder, email string
	flag.StringVar(&folder, "folder", "", "Path to the folder to scan")
	flag.StringVar(&email, "email", "", "Email address to get stats for")
	flag.Parse()

	if folder != "" {
		scan(folder)
	}

	stats(email)
}
