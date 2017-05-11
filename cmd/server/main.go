package main

import (
	"flag"

	"github.com/rollerderby/go/logger"
	"github.com/rollerderby/go/server"
)

func main() {
	port := flag.Int("port", 8000, "Port to listen on")
	verbose := flag.Bool("v", false, "Print debugging information")
	flag.Parse()

	if *verbose {
		logger.SetLevel(logger.DEBUG)
	}

	server.Run(uint16(*port))
}
