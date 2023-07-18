package main

import (
	"context"
	"honey_comb/handlers"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	uri, exists := os.LookupEnv("MONGODB_URI")
	if !exists {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable.")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	usersCollection := client.Database("test").Collection("users")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept",
	}))

	handlers := handlers.NewUsersHandler(usersCollection)

	app.Get("/user/all", handlers.GetUsers)
	app.Get("/user/:id", handlers.GetUser)
	app.Post("/user", handlers.CreateUser)
	app.Delete("/user/:id", handlers.DeleteUser)
	app.Post("/user/login", handlers.Login)

	app.Listen(":3000")
}
