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

	// For the REST API to accept app config files
	// "encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"bytes"
	"io"
	"strings"
)

// The struct holding the config data
type Config struct {
	App string
	Configs map[string]string
}

const POLL_FREQUENCY = 5 * time.Second
var cms pb.ConfigStoreClient

func ReceiveFile(w http.ResponseWriter, r *http.Request) {
	var Buf bytes.Buffer
	// in your case file would be fileupload
	file, header, err := r.FormFile("file")
	if err != nil {
			panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	fmt.Printf("File name %s\n", name[0])
	// Copy the file data to my buffer
	io.Copy(&Buf, file)

	config := Config{}
	contents := Buf.Bytes()

	errYaml := yaml.Unmarshal(contents, &config)
	if errYaml != nil {
		log.Fatalf("error: %v", errYaml)
	}
	
	configItems := []*pb.ConfigItem{}
	for configKey, configValue := range config.Configs {
		configItems = append(configItems, &pb.ConfigItem{Key: configKey, Value: configValue})
	}

	// Create a request for the key hello
	updateConfigReq := &pb.ConfigUpdateRequest{Application: config.App, Configs: configItems}
	// Send request to server.
	updateConfigRes, updateConfigErr := cms.Update(context.Background(), updateConfigReq)
	// Ensure request does not fail.
	if updateConfigErr != nil {
		log.Fatalf("Request error %v", updateConfigErr)
	}
	// Done
	log.Printf("Got response %v", updateConfigRes.Response)

	return
}

func usage() {
	fmt.Printf("Usage %s <endpoint> <config_yml>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	// Take endpoint as input
	flag.Usage = usage
	flag.Parse()
	// If there is no endpoint fail
	if flag.NArg() == 0 {
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

	// Create a ConfigStore client
	cms = pb.NewConfigStoreClient(conn)

	router := mux.NewRouter()
	router.HandleFunc("/v1/update", ReceiveFile).Methods("POST")
	log.Fatal(http.ListenAndServe(":8001", router))
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
