package main

import (
	"bytes"
	"encoding/json"
	// "flag"
	"fmt"
	"os"
	// "runtime"
	"time"

	"github.com/codegangsta/cli"

	_ "github.com/google/cayley/config"
	// "github.com/google/cayley/db"
	// "github.com/google/cayley/graph"
	// "github.com/google/cayley/http"
	// "github.com/google/cayley/internal"

	_ "github.com/google/cayley/graph/mongo"

	_ "github.com/google/cayley/writer"

	log "github.com/Sirupsen/logrus"
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

func loadConfigOn(c *cli.Context) (config *ogmaPrimeConfig, err error) {
	config, err = loadConfigFrom(c.GlobalString("config"))
	return
}

func loadConfigFrom(file string) (*ogmaPrimeConfig, error) {
	config := &ogmaPrimeConfig{}

	if file == "" {
		return config, nil
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Fatalf("Cannot find specified configuration file %q, aborting", file)
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

	setConfigDefaults(config)
	return config, nil
}

func setConfigDefaults(config *ogmaPrimeConfig) {
	if config.DatabaseType == "" {
		config.DatabaseType = "mongo"
	}

	if config.DatabasePath == "" {
		config.DatabasePath = "localhost:27017"
	}

	if config.ListenHost == "" {
		config.ListenHost = "0.0.0.0"
	}

	if config.ListenPort == "" {
		config.ListenPort = "22327"
	}
}

func main() {
	ogma := cli.NewApp()
	ogma.Name = "ogma"
	ogma.Version = Version

	ogma.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "config, c",
			Value: "./data/config.json",
			Usage: "Path to a configuration file",
			EnvVar: "OGMA_PRIME_CONFIG",
		},
		cli.BoolFlag{
			Name: "dump-profile",
			Usage: "Dump profiling information to a file",
		},
	}

	ogma.Commands = []cli.Command{
		{
			Name: "init",
			Usage: "Initialize and bootstrap",
		},
		{
			Name: "show-config",
			Usage: "Show configuration settings and exit",
			Action: showConfigAction,
		},
		{
			Name: "serve",
			Aliases: []string{"s", "srv"},
			Usage: "Serve HTTP",
			Action: serveAction,
		},
	}

	ogma.Run(os.Args)
}

func showConfigAction(c *cli.Context) {
	config, err := loadConfigOn(c)
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}

	dump, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Cannot dump configuration data: %v", err)
	}

	var buf bytes.Buffer
	json.Indent(&buf, dump, "", "  ")
	buf.WriteString("\n")
	buf.WriteTo(os.Stdout)
}

func serveAction(c *cli.Context) {
	println("test")
}
