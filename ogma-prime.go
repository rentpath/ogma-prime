package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	// "runtime"
	"time"

	"github.com/barakmich/glog"

	"github.com/codegangsta/cli"

	_ "github.com/google/cayley/config"
	// "github.com/google/cayley/db"
	// "github.com/google/cayley/graph"
	// "github.com/google/cayley/http"
	// "github.com/google/cayley/internal"

	_ "github.com/google/cayley/graph/mongo"

	_ "github.com/google/cayley/writer"
)

var (
	cpuProfile         = flag.String("prof", "", "Output profiling file.")
	configFile         = flag.String("config", "./config.json", "Path to a configuration file.")
)

// Filled in by `go build -ldflags="-X main.Version `ver`"`.
var (
	BuildDate string
	Version   string
)

type duration time.Duration

type ogmaPrimeConfig struct {
	DatabaseType string `json:"database_type"`
	DatabasePath string `json:"database_string"`
	ListenHost string `json:"listen_host"`
	ListenPort string `json:"listen_port"`
	Timeout duration `json:"timeout"`
}

func configFrom(file string) (*ogmaPrimeConfig, error) {
	if file != "" {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			glog.Fatalln("Cannot find specified configuration file", file, ", aborting.")
		}
	} else if _, err := os.Stat(os.Getenv("OGMA_PRIME_CONFIG")); err == nil {
		file = os.Getenv("OGMA_PRIME_CONFIG")
	}

	config := &ogmaPrimeConfig{}

	if file == "" {
		return config, nil
	}

	hnd, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %q: %v", file, err)
	}
	defer hnd.Close()

	decoder := json.NewDecoder(hnd)
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file %q: %v", file, err)
	}

	return config, nil
}

func main() {
	ogma := cli.NewApp()
	ogma.Name = "ogma"
	ogma.Version = Version
	ogma.Commands = []cli.Command{
		{
			Name: "serve",
			Aliases: []string{"s", "srv"},
			Usage: "Serve HTTP",
			Action: serveAction,
		},
	}
	ogma.Run(os.Args)
}

func serveAction(c *cli.Context) {
	println("test")
}
