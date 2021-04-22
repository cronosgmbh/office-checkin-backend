package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"time"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	FirebaseID string             `json:"firebase_id"`
	FirstName  string             `json:"first_name"`
	LastName   string             `json:"last_name"`
	Email      string             `json:"email"`
}

func getUser(c *gin.Context) {
	uid := c.GetString("userId")
	u, err := getSingleUser(uid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{
				Code:   http.StatusNotFound,
				Errors: []string{"no document found"},
			})
			return
		}
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	c.JSON(http.StatusOK, u)
}

func updateUser(c *gin.Context) {
	uid := c.GetString("userId")
	var u User
	if err := c.BindJSON(&u); err != nil {
		logrus.Info(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"body malformed"},
		})
		return
	}
	errs := []string{}
	if u.FirstName == "" {
		errs = append(errs, "first name cannot be empty")
	}
	if u.LastName == "" {
		errs = append(errs, "last name cannot be empty")
	}
	if u.Email == "" {
		errs = append(errs, "email cannot be empty")
	}
	u.FirebaseID = uid
	u.Email = c.GetString("userMail")
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}
	f := bson.D{{"firebase_id", uid}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	update := bson.M{
		"$set": struct {
			LastName   string `bson:"lastname"`
			FirstName  string `bson:"firstname"`
			Email      string `bson:"email"`
			FirebaseID string `bson:"firebaseid"`
		}{
			LastName:   u.LastName,
			FirstName:  u.FirstName,
			Email:      u.Email,
			FirebaseID: u.FirebaseID,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := client.Collection("users").UpdateMany(ctx, f, update, opts)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}

	c.Status(http.StatusOK)
}

func getSingleUser(firebaseID string) (u User, err error) {
	if firebaseID == "" {
		return User{}, errors.New("firebaseID cannot be empty")
	}

	f := bson.D{{"firebaseid", firebaseID}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = client.Collection("users").FindOne(ctx, f).Decode(&u)

	return
}
