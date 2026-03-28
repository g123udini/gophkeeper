package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	DBPath          string
	MasterPassword  string
	ServerAddr      string
	SyncIntervalSec int
}

func ParseArgs() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddr, "addr", "localhost:50501", "server address (host:port)")
	flag.IntVar(&cfg.SyncIntervalSec, "interval", 60, "synchronization interval in seconds")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [options] <db_path> <master_password>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()

		return nil, fmt.Errorf("invalid arguments")
	}

	cfg.DBPath = args[0]
	cfg.MasterPassword = args[1]

	return &cfg, nil
}
