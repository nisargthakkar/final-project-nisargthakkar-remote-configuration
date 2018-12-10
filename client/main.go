package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/nisargthakkar/final-project-nisargthakkar-remote-configuration/pb"

	// For the REST API between applications and sidecar
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
)

const POLL_FREQUENCY = 5 * time.Second

type ConfigItem struct {
	Key				string   `json:"key,omitempty"`
	Value 		string   `json:"value,omitempty"`
}

var configs map[string]string

func usage() {
	fmt.Printf("Usage %s <endpoint> <appname>\n", os.Args[0])
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

	appName := flag.Args()[1]

	configs = make(map[string]string)

	// Create a ConfigStore client
	rcs := pb.NewConfigStoreClient(conn)

	go poller(rcs, appName)

	router := mux.NewRouter()
	router.HandleFunc("/v1/config/{key}", GetConfig).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func GetConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	config := ConfigItem{Key: key, Value: ""}
	if val, ok := configs[key]; ok {
		config = ConfigItem{Key: key, Value: val}
	}

	json.NewEncoder(w).Encode(config)
}

func blockForever() {
	select{ }
}

func poller(rcs pb.ConfigStoreClient, appName string) {
	ticker := time.NewTicker(POLL_FREQUENCY)
	lastVersion := getConfig(rcs, appName, int64(0))

	if lastVersion == -1 {
		log.Fatalf("Unable to fetch configs")
	}

	for {
		select {
		case <- ticker.C:
			log.Printf("Timer went off. Getting updated configs")
			//Call the periodic function here.
			version := getConfig(rcs, appName, lastVersion)
			if version == -1 {
				continue
			}
			lastVersion = version
		}
	}
}

func getConfig(rcs pb.ConfigStoreClient, appName string, lastVersion int64) int64 {
	// Create a request for the key hello
	getConfigReq := &pb.ConfigRequest{Application: appName, PreviousVersion: lastVersion}
	// Send request to server.
	getConfigRes, getConfigErr := rcs.Get(context.Background(), getConfigReq)
	// Ensure request does not fail.
	if getConfigErr != nil {
		log.Fatalf("Request error %v", getConfigErr)
	}
	// Done
	log.Printf("Got response %v", getConfigRes.Response)

	if getConfigRes.GetErr() != nil {
		log.Printf("Got error: \"%v\"", getConfigRes.GetErr().Msg)
		return -1;
	}

	configUpdates := getConfigRes.GetConfig().Configs

	for _, configUpdate := range configUpdates {
		if configUpdate.Value == "" {
			delete(configs, configUpdate.Key);
		}
		configs[configUpdate.Key] = configUpdate.Value
		log.Printf("Updated app config: %v", configs)
	}

	return getConfigRes.GetConfig().Version
}
