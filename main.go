package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load .env file agar environment variable bisa dibaca di lokal/Docker
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, continuing...")
	}

	// Ambil URL Mongo dari environment
	mongoURI := os.Getenv("MONGO_PUBLIC_URL")
	mongoURI = strings.TrimSpace(mongoURI)
	mongoURI = strings.Trim(mongoURI, "\"") // buang tanda kutip kalau ada

	if mongoURI == "" {
		log.Fatal("❌ MONGO_PUBLIC_URL environment variable not set")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "gisdb" // default jika DB_NAME tidak di-set
	}

	// Koneksi ke MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("❌ Failed to connect MongoDB:", err)
	}
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal("❌ Cannot ping MongoDB:", err)
	}

	fmt.Println("✅ Connected to MongoDB successfully")

	db := client.Database(dbName)
	featuresCollection := db.Collection("features")

	r := gin.Default()

	// GET endpoint
	r.GET("/features", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "GET /features endpoint"})
	})

	// POST endpoint
	r.POST("/features", func(c *gin.Context) {
		var feature map[string]interface{}
		if err := c.BindJSON(&feature); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := featuresCollection.InsertOne(context.TODO(), feature)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Feature added successfully"})
	})

	r.Run(":8080")
}
