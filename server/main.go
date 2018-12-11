package main

import (
	"log"
	"net"
	"fmt"
	"strconv"
	"os"
	"flag"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/nisargthakkar/final-project-nisargthakkar-remote-configuration/pb"
)

// The struct holding the config-store data
type ConfigStore struct {}

type ConfigItem struct {
	Key string
	Value string
	Valid bool
	UpdateTime int64
}

const GET_LATEST_VERSION_QUERY = `SELECT c1.is_valid, c1.config_key, c1.config_value, c1.update_time
FROM (
	SELECT application, config_key, MAX(update_time) AS prev_max_time
	FROM configurations
	WHERE application = "%v"
	GROUP BY config_key
) AS c2
INNER JOIN configurations AS c1
ON c1.config_key = c2.config_key
AND c1.update_time = c2.prev_max_time
AND c1.application = c2.application`

const UPDATE_APP_CONFIG_QUERY = `UPDATE configurations
SET is_valid = True,
config_value = "%[3]v",
update_time = UNIX_TIMESTAMP(NOW())
WHERE config_key = "%[2]v"
AND application = "%[1]v";
`

const INSERT_APP_CONFIG_QUERY = `INSERT INTO configurations
(is_valid, application, config_key, config_value, update_time)
VALUES
(True, '%[1]v', '%[2]v', '%[3]v', UNIX_TIMESTAMP(NOW()));
`

const DELETE_APP_CONFIG_QUERY = `UPDATE configurations
SET is_valid = False,
config_value = "",
update_time = UNIX_TIMESTAMP(NOW())
WHERE config_key = "%[2]v"
AND application = "%[1]v";
`

const GET_UPDATES_QUERY = `SELECT c1.is_valid, c1.config_key, c1.config_value, c1.update_time
FROM (
	SELECT application, config_key, MAX(update_time) AS prev_max_time
	FROM configurations
	WHERE application = "%[1]v"
	GROUP BY config_key
) AS c2
INNER JOIN configurations AS c1
ON c1.config_key = c2.config_key
AND c1.update_time = c2.prev_max_time
AND c1.application = c2.application
WHERE c2.prev_max_time > %[2]v;`

const TRANSACTION_MARSHALLING_QUERY = `START TRANSACTION;
%v
COMMIT;`

// CREATE TABLE configurations (
// 	is_valid BOOLEAN,
// 	application VARCHAR(20), 
// 	config_key VARCHAR(20), 
// 	config_value TEXT, 
// 	update_time INT(11)
// );
var database *sql.DB 
var dbCtx = context.Background()

func usage() {
	fmt.Printf("Usage %s <mysql-endpoint>\n", os.Args[0])
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

	const Port = ":3000"
	// First initialize the store
	store := ConfigStore{}

	mysqlconfig := fmt.Sprintf("root:distributed_systems@(%v)/appconfig", endpoint)
	db, err := sql.Open("mysql", mysqlconfig)

	if err != nil {
		log.Fatalf("Failed to connect to MySQL server")
	}

	database = db

	// Create socket that listens on port 3000
	c, err := net.Listen("tcp", Port)
	if err != nil {
		// Note the use of Fatalf which will exit the program after reporting the error.
		log.Fatalf("Could not create listening socket %v", err)
	}
	// Create a new GRPC server
	s := grpc.NewServer()
	// Tell GRPC that s will be serving requests for the ConfigStore service and should use store (defined on line 23)
	// as the struct whose methods should be called in response.
	pb.RegisterConfigStoreServer(s, &store)
	log.Printf("Going to listen on port %v", Port)
	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
	log.Printf("Done listening")
}

// Handle the Get RPC from the remote config client sidecar
func (s *ConfigStore) Get(ctx context.Context, req *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	// The bit below works because Go maps return the 0 value for non existent keys, which is empty in this case.
	appName := req.Application
	previousTimestamp := req.PreviousVersion

	configUpdates := getConfigUpdates(appName, previousTimestamp)

	maxTimestamp := previousTimestamp
	configItems := []*pb.ConfigItem{}
	for key, configItem := range configUpdates {
		configItems = append(configItems, &pb.ConfigItem{Key: key, Value: configItem.Value})
		if maxTimestamp < configItem.UpdateTime {
			maxTimestamp = configItem.UpdateTime
		}
	}

	config := &pb.Config{Version: maxTimestamp, Configs: configItems}
	return &pb.ConfigResponse{Response: &pb.ConfigResponse_Config{Config: config}}, nil
}

func getConfigUpdates(appName string, previousTimestamp int64) map[string]ConfigItem {
	sqlQuery := fmt.Sprintf(GET_UPDATES_QUERY, appName, previousTimestamp)
	stmt, err := database.Prepare(sqlQuery)
	
	currentAppConfig := make(map[string]ConfigItem)

	if err != nil {
			log.Fatal(err)
	}
	defer stmt.Close() // closing the statement
	rows, err := stmt.Query()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		isValid := string(values[0])[0] != '0'
		key := string(values[1])
		value := string(values[2])
		updateTime, parseErr := strconv.ParseInt(string(values[3]), 10, 64)

		if parseErr != nil {
			log.Fatalf("Unable to parse update time: %v to int64", string(values[3]))
		}

		currentAppConfig[key] = ConfigItem{
			Key: key,
			Value: value,
			Valid: isValid,
			UpdateTime: updateTime,
		}
	}

	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	return currentAppConfig
}

func getCurrentConfig(appName string) map[string]ConfigItem {
	sqlQuery := fmt.Sprintf(GET_LATEST_VERSION_QUERY, appName)
	stmt, err := database.Prepare(sqlQuery)
	
	currentAppConfig := make(map[string]ConfigItem)

	if err != nil {
			log.Fatal(err)
	}
	defer stmt.Close() // closing the statement
	rows, err := stmt.Query()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		isValid := string(values[0])[0] != '0'
		key := string(values[1])
		value := string(values[2])
		// updateTime := string(values[3])

		currentAppConfig[key] = ConfigItem{
			Key: key,
			Value: value,
			Valid: isValid,
		}
	}

	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	return currentAppConfig
}

func executeSqlQuery(sqlQuery string) bool {
	stmt, err := database.Prepare(sqlQuery)

	if err != nil {
			log.Fatal(err)
	}
	defer stmt.Close() // closing the statement
	rows, err := stmt.Query()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	log.Printf("%v", columns)

	return true
}

func executeSqlQueriesInTransaction(sqlQueries []string) error {
	if len(sqlQueries) == 0 {
		return nil
	}

	tx, err := database.BeginTx(dbCtx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	
	for _, sqlQuery := range sqlQueries {
		_, execErr := tx.Exec(sqlQuery)
		if execErr != nil {
			_ = tx.Rollback()
			return execErr
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *ConfigStore) Update(ctx context.Context, req *pb.ConfigUpdateRequest) (*pb.ConfigUpdateResponse, error) {
	appName := req.Application

	currentAppConfig := getCurrentConfig(appName)
	updateAppConfig := make(map[string]string)
	insertAppConfig := make(map[string]string)
	invalidAppConfig := make([]string, 0)
	requestedAppConfig := make(map[string]string)

	for _, config := range req.Configs {
		requestedAppConfig[config.Key] = config.Value
		if _, ok := currentAppConfig[config.Key]; !ok {
			insertAppConfig[config.Key] = config.Value
		}
	}

	for key, value := range currentAppConfig {
		if updateConfigValue, ok := requestedAppConfig[key]; ok {
			if value.Value != updateConfigValue || !value.Valid {
				updateAppConfig[key] = updateConfigValue
			}
		} else if value.Valid {
			invalidAppConfig = append(invalidAppConfig, key)
		}
	}

	sqlQueries := make([]string, 0)
	for key, value := range updateAppConfig {
		thisQuery := fmt.Sprintf(UPDATE_APP_CONFIG_QUERY, appName, key, value)
		sqlQueries = append(sqlQueries, thisQuery)
	}

	for _, key := range invalidAppConfig {
		thisQuery := fmt.Sprintf(DELETE_APP_CONFIG_QUERY, appName, key)
		sqlQueries = append(sqlQueries, thisQuery)
	}

	for key, value := range insertAppConfig {
		thisQuery := fmt.Sprintf(INSERT_APP_CONFIG_QUERY, appName, key, value)
		sqlQueries = append(sqlQueries, thisQuery)
	}

	err := executeSqlQueriesInTransaction(sqlQueries)

	if err != nil {
		return &pb.ConfigUpdateResponse{Response: &pb.ConfigUpdateResponse_Err{Err: &pb.Error{Msg: err.Error()}}}, nil
	}

	return &pb.ConfigUpdateResponse{Response: &pb.ConfigUpdateResponse_Success{Success: &pb.Success{}}}, nil
}