package api

import (
	"os"
	"fmt"
	"github.com/gocql/gocql"
)

var Session *gocql.Session
var KEYSPACE_NAME = "engine"

func init() {
	var err error 

	host, isSet := os.LookupEnv("DB_HOST")
	if !isSet {
		host = "127.0.0.1"
	}
	
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = KEYSPACE_NAME
	Session, err = cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	fmt.Println("cassandra well initialized")
}