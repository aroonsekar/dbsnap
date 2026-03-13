//go:build integration
// +build integration

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMongoContainer(t *testing.T) (testcontainers.Container, string) {
	// Recover from testcontainers panicking on some Windows Docker setups
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping integration tests due to local Docker environment limitation: %v", r)
		}
	}()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Skipping integration test, could not start container: %v", err)
	}

	endpoint, err := mongoC.Endpoint(ctx, "")
	require.NoError(t, err)

	uri := "mongodb://" + endpoint
	return mongoC, uri
}

func TestMongoClientIntegration(t *testing.T) {
	container, uri := setupMongoContainer(t)
	defer container.Terminate(context.Background())

	t.Run("Connect successfully", func(t *testing.T) {
		client, err := Connect(uri)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("ListDatabases ignores system databases", func(t *testing.T) {
		client, err := Connect(uri)
		require.NoError(t, err)

		// Create a test database and collection to ensure it drops something
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = client.Database("test_db").Collection("dummy").InsertOne(ctx, map[string]string{"foo": "bar"})
		require.NoError(t, err)

		dbs, err := ListDatabases(client)
		require.NoError(t, err)
		assert.Contains(t, dbs, "test_db")
		assert.NotContains(t, dbs, "admin")
		assert.NotContains(t, dbs, "local")
	})

	t.Run("ListCollections and PurgeCollection", func(t *testing.T) {
		client, err := Connect(uri)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dbName := "integration_db"
		collName := "integration_coll"
		
		// Insert documents to create collection
		_, err = client.Database(dbName).Collection(collName).InsertMany(ctx, []interface{}{
			map[string]string{"foo": "bar"},
			map[string]string{"foo": "baz"},
		})
		require.NoError(t, err)

		// List collections
		colls, err := ListCollections(client, dbName)
		require.NoError(t, err)
		assert.Contains(t, colls, collName)

		// Purge collection
		count, err := PurgeCollection(client, dbName, collName)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Verify empty
		countResult, err := client.Database(dbName).Collection(collName).CountDocuments(ctx, map[string]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), countResult)
	})
}
