package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"io/ioutil"
	"gopkg.in/yaml.v2"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/nisargthakkar/final-project-nisargthakkar-remote-configuration/pb"
)

// The struct holding the config data
type Config struct {
	App string
	Configs map[string]string
}

const POLL_FREQUENCY = 5 * time.Second

func usage() {
	fmt.Printf("Usage %s <endpoint> <config_yml>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	// Take endpoint as input
	flag.Usage = usage
	flag.Parse()
	// If there is no endpoint fail
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}
	endpoint := flag.Args()[0]
	log.Printf("Connecting to %v", endpoint)
	// Connect to the server. We use WithInsecure since we do not configure https in this class.
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	//Ensure connection did not fail.
	if err != nil {
		log.Fatalf("Failed to dial GRPC server %v", err)
	}
	log.Printf("Connected")

	// Create a KvStore client
	rcs := pb.NewConfigStoreClient(conn)
	
	config := readConfigFromFile(flag.Args()[1])

	configItems := []*pb.ConfigItem{}
	for configKey, configValue := range config.Configs {
		configItems = append(configItems, &pb.ConfigItem{Key: configKey, Value: configValue})
	}

	// Create a request for the key hello
	updateConfigReq := &pb.ConfigUpdateRequest{Application: config.App, Configs: configItems}
	// Send request to server.
	updateConfigRes, updateConfigErr := rcs.Update(context.Background(), updateConfigReq)
	// Ensure request does not fail.
	if updateConfigErr != nil {
		log.Fatalf("Request error %v", updateConfigErr)
	}
	// Done
	log.Printf("Got response %v", updateConfigRes.Response)
}

func blockForever() {
	select{ }
}

func readConfigFromFile(filename string) Config {
	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatalf("Error reading file")
	}

	config := Config{}

	errYaml := yaml.Unmarshal(contents, &config)
	if errYaml != nil {
		log.Fatalf("error: %v", errYaml)
	}

	return config
}
