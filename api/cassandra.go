package api

import (
	"fmt"
	"github.com/gocql/gocql"
)

var Session *gocql.Session
var KEYSPACE_NAME = "engine"

func init() {
	var err error 
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = KEYSPACE_NAME
	Session, err = cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	fmt.Println("cassandra well initialized")
}