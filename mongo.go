package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var (
	mongoClient *mongo.Client
	collection  *mongo.Collection
)

var (
	GetMongoSecret = getMongoSecret
	AddRowToMongo  = addRowToMongo
)

// MongoEntry represents a reading list entry in MongoDB
type MongoEntry struct {
	URL           string    `bson:"url"`
	Title         string    `bson:"title"`
	Description   string    `bson:"description"`
	Image         string    `bson:"image,omitempty"`
	Date          time.Time `bson:"date"`
	HackerNewsURL string    `bson:"hackerNewsUrl,omitempty"`
	Screenshot    string    `bson:"screenshot,omitempty"`
	PDF           string    `bson:"pdf,omitempty"`
	Domain        string    `bson:"domain"`
}

// connectToDB connects to a mongo DB instance and initializes the global client
func connectToDB(secret string) error {
	if mongoClient != nil {
		return nil // Already connected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := fmt.Sprintf("mongodb+srv://suryaatvellore:%s@clusterrl.sa6ipg8.mongodb.net/?retryWrites=true&w=majority&appName=ClusterRL", secret)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	mongoClient = client
	collection = client.Database("readingList").Collection("entries")
	fmt.Println("Successfully connected to MongoDB!")
	return nil
}

// disconnectDB disconnects from MongoDB
func disconnectDB() error {
	if mongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mongoClient.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	mongoClient = nil
	collection = nil
	return nil
}

// insertEntry adds a new entry to MongoDB
func insertEntry(entry *MongoEntry) error {
	if mongoClient == nil {
		return fmt.Errorf("not connected to MongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	return nil
}

// getMongoSecret ...
func getMongoSecret() (string, error) {
	mongoSecret := os.Getenv("MONGO_CONN_TOKEN")
	if mongoSecret == "" {
		return "", fmt.Errorf("MONGO_CONN_TOKEN environment variable not set")
	}
	return mongoSecret, nil
}

func addRowToMongo(data *readingListEntry, hnURL, domain string) error {
	secret, err := GetMongoSecret()
	if err != nil {
		return fmt.Errorf("failed to get MongoDB secret: %w", err)
	}
	// Connect to MongoDB
	if err := connectToDB(secret); err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer disconnectDB()

	// Create MongoDB entry
	mongoEntry := &MongoEntry{
		URL:           data.URL,
		Title:         data.Title,
		Description:   strings.ReplaceAll(data.Description, "\n", " "),
		Image:         data.Image,
		Date:          time.Now(),
		HackerNewsURL: hnURL,
		Screenshot:    data.Screenshot,
		PDF:           data.PDF,
		Domain:        domain,
	}

	// Insert into MongoDB
	if err := insertEntry(mongoEntry); err != nil {
		return fmt.Errorf("failed to insert into MongoDB: %w", err)
	}
	return nil
}
