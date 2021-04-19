package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type Settings struct {
	ID primitive.ObjectID `bson:"_id" ,json:"id"`
	Key string `bson:"key" ,json:"key"`
	LocationManagers []string `bson:"location_managers" ,json:"location_managers"`
}

var (
	adminUsers map[string]bool
)

func initSettings() {
	logrus.Info("init general settings")
	f := bson.D{{"key", "general_settings"}}
	var s Settings
	err := client.Database("office_checkin").Collection("settings").FindOne(context.Background(), f).Decode(&s)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("found %d admin users", len(s.LocationManagers))
	adminUsers = make(map[string]bool, len(s.LocationManagers))
	for _,user := range s.LocationManagers {
		adminUsers[user] = true
		logrus.WithField("user_mail", user).Debug("adding location manager")
	}
}

func refreshSettingsHandler(c *gin.Context) {
	initSettings()
	c.JSON(http.StatusOK, nil)
}