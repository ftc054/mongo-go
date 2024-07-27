package main

import (
	"context"
	"fmt"
	"mongodb-practice/config"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var client *mongo.Client
var customerCollection *mongo.Collection
var productCollection *mongo.Collection
var orderCollection *mongo.Collection

func main() {
	// Setup zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to set up zap logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() // Flushes buffer, if any

	config.LoadEnv()

	mongoURI := config.GetMongoDB_URL()
	if mongoURI == "" {
		logger.Fatal("Environment variable for MongoDB URL is not set.")
	}

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	// Create a client and connect to the server
	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB.", zap.Error(err))
	}

	// Schedule a deferred disconnection
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			logger.Error("Failed to disconnect from MongoDB.", zap.Error(err))
		}
	}()

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send a ping command to confirm connection
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		logger.Fatal("Failed to ping MongoDB.", zap.Error(err))
	}

	logger.Info("Successfully connected to MongoDB.")

	customerCollection = client.Database("firstDB").Collection("customers")
	productCollection = client.Database("firstDB").Collection("products")
	orderCollection = client.Database("firstDB").Collection("orders")

	app := fiber.New()

	// Middleware
	app.Use(fiberLogger.New())

	// Serve static files

	app.Get("/api/fields", listFields)

	app.Get("/api/customers", getAllCustomers)
	app.Get("/api/customers/:id", getCustomerByID)
	app.Get("/api/products", getAllProducts)
	app.Get("/api/products/:id", getProductByID)
	app.Get("/api/orders", getAllOrders)
	app.Get("/api/orders/:id", getOrderByID)

	if err := app.Listen(":8090"); err != nil {
		logger.Fatal("Failed to start server.", zap.Error(err))
	}
}

func listFields(c *fiber.Ctx) error {
	// Get a list of all collection names in the database
	collections, err := client.Database("firstDB").ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	fields := make(map[string][]string)

	for _, collection := range collections {
		cursor, err := client.Database("firstDB").Collection(collection).Find(context.Background(), bson.D{})
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer cursor.Close(context.Background())

		var result bson.M
		if cursor.Next(context.Background()) {
			if err := cursor.Decode(&result); err != nil {
				return c.Status(500).SendString(err.Error())
			}
			for key := range result {
				fields[collection] = append(fields[collection], key)
			}
		}
	}

	return c.JSON(fields)
}

func getAllCustomers(c *fiber.Ctx) error {
	cursor, err := customerCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer cursor.Close(context.Background())

	var customers []bson.M
	if err = cursor.All(context.Background(), &customers); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(customers)
}

func getCustomerByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var customer bson.M

	// Use the ObjectID from the provided ID string
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).SendString("Invalid ID format")
	}

	filter := bson.M{"_id": objectID}

	if err := customerCollection.FindOne(context.Background(), filter).Decode(&customer); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(404).SendString("Customer not found")
		}
		return c.Status(500).SendString("Error finding customer: " + err.Error())
	}

	return c.JSON(customer)
}

func getAllProducts(c *fiber.Ctx) error {
	cursor, err := productCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer cursor.Close(context.Background())

	var products []bson.M
	if err = cursor.All(context.Background(), &products); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(products)
}

func getProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var product bson.M

	// Use the ObjectID from the provided ID string
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).SendString("Invalid ID format")
	}

	filter := bson.M{"_id": objectID}

	if err := productCollection.FindOne(context.Background(), filter).Decode(&product); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(404).SendString("Product not found")
		}
		return c.Status(500).SendString("Error finding product: " + err.Error())
	}

	return c.JSON(product)
}

func getAllOrders(c *fiber.Ctx) error {
	cursor, err := orderCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer cursor.Close(context.Background())

	var orders []bson.M
	if err = cursor.All(context.Background(), &orders); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(orders)
}

func getOrderByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var order bson.M

	// Use the ObjectID from the provided ID string
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).SendString("Invalid ID format")
	}

	filter := bson.M{"_id": objectID}

	if err := orderCollection.FindOne(context.Background(), filter).Decode(&order); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(404).SendString("Order not found")
		}
		return c.Status(500).SendString("Error finding order: " + err.Error())
	}

	return c.JSON(order)
}
