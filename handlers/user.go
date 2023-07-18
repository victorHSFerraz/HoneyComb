package handlers

import (
	"context"
	"honey_comb/models"
	"honey_comb/utils"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UsersHandler struct {
	collection *mongo.Collection
}

func NewUsersHandler(collection *mongo.Collection) *UsersHandler {
	return &UsersHandler{collection: collection}
}

func (h *UsersHandler) GetUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var users []models.User
	cursor, err := h.collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var user models.User
		cursor.Decode(&user)
		users = append(users, user)
	}

	if len(users) == 0 {
		return c.JSON(fiber.Map{"message": "No users found"})
	}

	return c.JSON(users)
}

func (h *UsersHandler) GetUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	id, _ := primitive.ObjectIDFromHex(c.Params("id"))
	user := &models.User{}
	if err := h.collection.FindOne(ctx, bson.M{"_id": id}).Decode(user); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(user)
}

func (h *UsersHandler) CreateUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	var existingUser models.User
	err := h.collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}
	} else {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "User already exists"})
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	user.Password = hashedPassword

	res, err := h.collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	user.ID = res.InsertedID.(primitive.ObjectID)
	user.Password = ""
	return c.Status(http.StatusCreated).JSON(user)
}

func (h *UsersHandler) DeleteUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	res, err := h.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	if res.DeletedCount < 1 {
		return c.Status(http.StatusNotFound).SendString("User not found")
	}

	return c.JSON(fiber.Map{"message": "User deleted"})
}

func (h *UsersHandler) Login(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	var foundUser models.User
	err := h.collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid email or password"})
	}

	if !utils.CheckPassword(user.Password, foundUser.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid email or password"})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = foundUser.ID
	claims["email"] = foundUser.Email
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	secretKey := os.Getenv("SECRET_KEY")
	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{
			"tokens": t,
		},
	}
	_, err = h.collection.UpdateOne(ctx, bson.M{"_id": foundUser.ID}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(fiber.Map{"token": t})
}
