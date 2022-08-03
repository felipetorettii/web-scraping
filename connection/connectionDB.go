package connection

import (
	"fmt"
	"os"

	"github.com/Junkes887/artifacts"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// ConnectionDB create connection with database
func ConnectionDB() (driver neo4j.Driver) {
	dbUri := os.Getenv("DB_URL")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth(dbUser, dbPassword, ""))
	artifacts.HandlerError(err)

	fmt.Println("Connection Database")

	return driver
}
