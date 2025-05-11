package main

import (
	"flag"
)

func main() {
	var folder, email string
	flag.StringVar(&folder, "folder", "", "Path to the folder to scan")
	flag.StringVar(&email, "email", "", "Email address to get stats for")
	flag.Parse()

	if folder != "" {
		scan(folder)
	}

	stats(email)
}
