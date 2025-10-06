package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func initMongo() {
	// Load env
	_ = godotenv.Load()

	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	colName := os.Getenv("MONGO_COLLECTION")

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database(dbName).Collection(colName)
	fmt.Println("✅ Connected to MongoDB successfully")
}

// Struct GeoJSON Feature
type Feature struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type       string             `bson:"type" json:"type"`
	Geometry   interface{}        `bson:"geometry" json:"geometry"`
	Properties interface{}        `bson:"properties" json:"properties"`
}

func main() {
	initMongo()
	router := gin.Default()

	// GET all features
	router.GET("/features", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var features []Feature
		if err = cursor.All(ctx, &features); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, features)
	})

	// POST new feature
	router.POST("/features", func(c *gin.Context) {
		var feature Feature
		if err := c.BindJSON(&feature); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		feature.Type = "Feature"

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := collection.InsertOne(ctx, feature)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"insertedID": res.InsertedID})
	})

	// PUT (update)
	router.PUT("/features/:id", func(c *gin.Context) {
		id, _ := primitive.ObjectIDFromHex(c.Param("id"))
		var feature Feature
		if err := c.BindJSON(&feature); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{"$set": feature}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Feature updated"})
	})

	// DELETE feature
	router.DELETE("/features/:id", func(c *gin.Context) {
		id, _ := primitive.ObjectIDFromHex(c.Param("id"))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Feature deleted"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}