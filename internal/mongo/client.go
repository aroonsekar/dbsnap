package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect establishes a verified connection to the cluster with a strict 10-second timeout.
func Connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %v", err)
	}

	// Force a ping to verify the network connection is actually alive
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("database unreachable (check network or IP whitelist): %v", err)
	}

	return client, nil
}

// PurgeCollection safely deletes all documents while preserving indexes and schema.
func PurgeCollection(client *mongo.Client, dbName, collName string) (int64, error) {
	collection := client.Database(dbName).Collection(collName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := collection.DeleteMany(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// ListDatabases fetches all non-system databases from the cluster.
func ListDatabases(client *mongo.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	names, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, name := range names {
		// Filter out standard MongoDB system databases
		if name != "admin" && name != "config" && name != "local" && name != "macrometa" {
			filtered = append(filtered, name)
		}
	}
	return filtered, nil
}

// ListCollections fetches all collections for a specific database.
func ListCollections(client *mongo.Client, dbName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	names, err := client.Database(dbName).ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	return names, nil
}
