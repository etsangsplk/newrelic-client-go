package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"gopkg.in/yaml.v2"

	"github.com/newrelic/newrelic-client-go/internal/http"
	nrConfig "github.com/newrelic/newrelic-client-go/pkg/config"

	log "github.com/sirupsen/logrus"
)

// Config is the information keeper for generating go structs from type names.
type Config struct {
	Package string   `yaml:"package"`
	Types   []string `yaml:"types"`
}

func main() {
	var (
		config Config
	)

	verbose := flag.Bool("v", false, "increase verbosity")
	flag.StringVar(&config.Package, "p", "", "package name")

	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	nrCfg := nrConfig.New()
	nrCfg.PersonalAPIKey = os.Getenv("NEW_RELIC_API_KEY")

	nrClient := http.NewClient(nrCfg)

	schemaResponse := allTypesResponse{}
	vars := map[string]interface{}{}
	err := nrClient.NerdGraphQuery(allTypes, vars, &schemaResponse)
	if err != nil {
		log.Fatal(err)
	}
	schema := &schemaResponse.Schema

	yamlFile, err := ioutil.ReadFile("typegen.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	types, err := ResolveSchemaTypes(*schema, config.Types)
	if err != nil {
		log.Error(err)
	}

	f, err := os.Create("types.go")
	if err != nil {
		log.Error(err)
	} else {

		_, err = f.WriteString(fmt.Sprintf("// Code generated by typegen; DO NOT EDIT.\n\npackage %s\n", config.Package))
		if err != nil {
			log.Error(err)
		}

		defer f.Close()

		keys := make([]string, 0, len(types))
		for k := range types {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			_, err := f.WriteString(types[k])
			if err != nil {
				log.Error(err)
			}
		}
	}
}
