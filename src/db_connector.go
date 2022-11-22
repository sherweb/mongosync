package src

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type DBConnector struct {
	SourceURI string
	DestURI   string
	srcConn   *mongo.Client
	destConn  *mongo.Client
}
