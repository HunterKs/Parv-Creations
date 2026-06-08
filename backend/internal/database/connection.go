package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	databaseName    = "parv_creation"
	defaultTimeout  = 10 * time.Second
	maxPoolSize     = 100
	minPoolSize     = 10
	maxConnIdleTime = 5 * time.Minute
)

// ConnectProductionCluster reads MONGO_PRODUCTION_URI from the environment, initializes
// a MongoDB v2 client with a high-throughput pool, verifies connectivity, and returns
// both the client and the selected database handle.
func ConnectProductionCluster() (*mongo.Client, *mongo.Database, error) {
	// 1. Enforce Zero-Leak environment guardrail by checking system shell values first
	mongoURI := os.Getenv("MONGO_PRODUCTION_URI")
	if mongoURI == "" {
		return nil, nil, fmt.Errorf("missing required environment variable MONGO_PRODUCTION_URI")
	}

	// 2. Establish a life-jacket context to abort stalled network connections after 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 3. Apply high-throughput connection pool scaling thresholds
	clientOptions := options.Client().ApplyURI(mongoURI).
		SetMaxPoolSize(100). // Maintain up to 100 simultaneous open sockets for concurrency
		SetMinPoolSize(10).  // Keep 10 warm idle connections ready to instantly serve queries
		SetMaxConnIdleTime(5 * time.Minute)

	// 4. Initialize connection runtime via the explicit MongoDB v2 framework
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect failed: %w", err)
	}

	// 5. Fire a low-overhead ping frame to confirm the cloud Atlas connection is active
	err = client.Ping(ctx, nil)
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("mongo ping failed: %w", err)
	}

	// 6. Bind database reference to the targeted production namespace
	db := client.Database(databaseName)
	return client, db, nil
}
