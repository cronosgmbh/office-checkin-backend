package main

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var authClient *auth.Client

func initFirebase() {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		logrus.Fatal(err)
	}
	authClient, err = app.Auth(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
}

func authMiddleware () gin.HandlerFunc {
	return func (c *gin.Context) {
		logrus.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path": c.Request.URL.Path,
			"host": c.Request.Host,
		}).Info("handling request")

		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Token ") {
			authHeader = strings.TrimPrefix(authHeader, "Token ")
		}
		token, err := authClient.VerifyIDToken(context.Background(), authHeader)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorTokenInvalidOrNotFound)
			return
		}
		c.Set("userId", token.UID)
		user, err := authClient.GetUser(context.Background(), token.UID)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorTokenInvalidOrNotFound)
			return
		}

		if !strings.HasSuffix(user.Email, "@cronos.de") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Code:   http.StatusUnauthorized,
				Errors: []string{
					"you are not allowed to access this endpoint",
					"you have to be logged in with a cronos mail address",
				},
			})
			return
		}

		if !adminUsers[user.Email] {
			c.Set("isAdmin", false)
		} else {
			c.Set("isAdmin", true)
		}

		c.Set("userMail", user.Email)
		c.Set("userDisplayName", user.DisplayName)
		c.Set("userRecord", user)
	}
}


func corsHeader() gin.HandlerFunc {
	return func (c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}
	}
}