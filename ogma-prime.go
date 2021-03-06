package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/cli"

	cayleyConfig "github.com/google/cayley/config"
	cayleyDb "github.com/google/cayley/db"
	cayleyGraph "github.com/google/cayley/graph"
	cayleyGremlin "github.com/google/cayley/query/gremlin"
	cayleyQuery "github.com/google/cayley/query"

	_ "github.com/google/cayley/graph/mongo"
	_ "github.com/google/cayley/writer"

	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"

	"gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
)

// Filled in by `go build -ldflags="-X main.Version `ver`"`.
var (
	BuildDate string
	Version   string
)

type duration time.Duration

type ogmaPrimeConfig struct {
	DatabaseType string   `json:"database_type"`
	DatabasePath string   `json:"database_string"`
	ListenHost   string   `json:"listen_host"`
	ListenPort   string   `json:"listen_port"`
	Timeout      duration `json:"timeout"`
}

func (config *ogmaPrimeConfig) CayleyConfig() (cayley *cayleyConfig.Config) {
	cayley = &cayleyConfig.Config{
		DatabaseType:    config.DatabaseType,
		DatabasePath:    config.DatabasePath,
		ReplicationType: "single",
		Timeout:         time.Duration(20 * time.Second),
	}

	return
}

func loadConfigOn(c *cli.Context) (config *ogmaPrimeConfig) {
	config, err := loadConfigFrom(c.GlobalString("config"))
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}

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
			Name:   "config, c",
			Value:  "./data/config.json",
			Usage:  "Path to a configuration file",
			EnvVar: "OGMA_PRIME_CONFIG",
		},
		cli.BoolFlag{
			Name:  "dump-profile",
			Usage: "Dump profiling information to a file",
		},
	}

	ogma.Commands = []cli.Command{
		{
			Name:   "dump",
			Usage:  "Dump Cayley contents",
			Action: dumpAction,
		},
		{
			Name:   "init",
			Usage:  "Bootstrap and initialize",
			Action: initAction,
		},
		{
			Name:   "show-config",
			Usage:  "Show configuration settings and exit",
			Action: showConfigAction,
		},
		{
			Name:    "serve",
			Aliases: []string{"s", "srv"},
			Usage:   "Serve HTTP",
			Action:  serveAction,
		},
	}

	ogma.Run(os.Args)
}

func initAction(c *cli.Context) {
	config := loadConfigOn(c)
	err := cayleyDb.Init(config.CayleyConfig())
	if err != nil {
		log.Fatalf("Could not bootstrap database: %v", err)
	}
}

func mongoSession(config *ogmaPrimeConfig) (session *mgo.Session, err error) {
	session, err = mgo.Dial(config.DatabasePath)
	return
}

func mongoShow(config *ogmaPrimeConfig) {
	session, err := mongoSession(config)
	if err != nil {
		log.Fatalf("Cannot connect to MongoDB: %v", err)
	}

	// dbs, err := session.DatabaseNames()
	// if err != nil {
	// 	log.Fatalf("Cannot retrieve database names: %v\n", err)
	// }

	// fmt.Printf("Available databases are: %v\n", dbs)

	db := session.DB("cayley")
	collection := db.C("quads")
	query := collection.Find(nil)

	count, err := query.Count()
	if err != nil {
		log.Errorf("Cannot count result set: %v", err)
	}
	fmt.Printf("Found %d results:\n", count)

	cursor := query.Iter()

	var result *interface{}
	for cursor.Next(&result) {
		fmt.Printf("%v\n", result)
	}

	if err = cursor.Close(); err != nil {
		log.Fatalf("Cannot close cursor: %v", err)
	}
}

func dumpAction(c *cli.Context) {
	config := loadConfigOn(c)
	mongoShow(config)
}

func showConfigAction(c *cli.Context) {
	config := loadConfigOn(c)
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
	config := loadConfigOn(c)
	cayConfig := config.CayleyConfig()

	graph, err := cayleyDb.Open(cayConfig)
	if err != nil {
		log.Fatalf("Cannot open database: %v", err)
	}

	http.Handle("/", serveInstallRoutes(graph, cayConfig))

	log.Infof("Listening on %s:%s", config.ListenHost, config.ListenPort)
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", config.ListenHost, config.ListenPort), nil)
	if err != nil {
		log.Fatalf("Cannot listen and serve on %s:%s: %v", config.ListenHost, config.ListenPort, err)
	}
}

func serveInstallApiV1(router *mux.Router, graph *cayleyGraph.Handle) {
	router.HandleFunc("/properties/{id}", apiLogger(graph, findProperty)).Methods("GET")
}

func serveInstallRoutes(graph *cayleyGraph.Handle, config *cayleyConfig.Config) *mux.Router {
	root := mux.NewRouter()
	serveInstallApiV1(root.PathPrefix("/api/v1").Subrouter(), graph)

	return root
}

type apiHandler func(http.ResponseWriter, *http.Request)
type graphHandler func(graph *cayleyGraph.Handle, rsp http.ResponseWriter, req *http.Request) int

func apiLogger(graph *cayleyGraph.Handle, handler graphHandler) apiHandler {
	return func(rsp http.ResponseWriter, req *http.Request) {
		start := time.Now()
		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}

		retcode := handler(graph, rsp, req)
		log.Infof("%s %s %d %v %s", req.Method, addr, retcode, time.Since(start), req.URL.Path)
	}
}

func findProperty(graph *cayleyGraph.Handle, rsp http.ResponseWriter, req *http.Request) int {
	result := make(map[string]interface{})
	retcode := 200

	pathVars := mux.Vars(req)
	result["request"] = pathVars

	session := cayleyGremlin.NewSession(graph.QuadStore, time.Duration(10 * time.Second), false)
	gremlinQuery := fmt.Sprintf(`g.V("/properties/%s").All()`, pathVars["id"])

	queryResult, err := session.Parse(gremlinQuery)
	switch queryResult {
	case cayleyQuery.Parsed:
		output, err := runGremlinQuery(gremlinQuery, session)
		if err != nil {
			result["success"] = false
			result["error"] = err
			retcode = 400
			break
		}

		result["success"] = true
		result["output"] = output
	case cayleyQuery.ParseFail:
		result["success"] = false
		result["error"] = fmt.Sprintf("Failed to parse query: %s", err)
		retcode = 400
	default:
		result["success"] = false
		result["error"] = "Possibly incomplete data or query?"
		retcode = 500
	}

	bytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintln(rsp, string(bytes))
	return retcode
}

func runGremlinQuery(q string, session cayleyQuery.HTTP) (interface{}, error) {
	c := make(chan interface{}, 5)
	go session.Execute(q, c, 100)
	for result := range c {
		session.Collate(result)
	}

	return session.Results()
}
