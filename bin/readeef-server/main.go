package main

import (
	"flag"
	"fmt"
	"os"
	"readeef"
	"readeef/api"

	"github.com/urandom/webfw"
)

func main() {
	confpath := flag.String("config", "", "config path")
	host := flag.String("host", "", "server host")
	port := flag.Int("port", 0, "server port")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*confpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *confpath, err))
	}

	server := webfw.NewServer(*confpath)
	if *host != "" {
		server.SetHost(*host)
	}

	if *port > 0 {
		server.SetPort(*port)
	}

	dispatcher := server.Dispatcher("/api/")

	api.RegisterControllers(cfg, dispatcher)
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}
