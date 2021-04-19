package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strings"
	"time"
)

func covidBacktracing(c *gin.Context) {

	if !c.GetBool("isAdmin") {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrorForbidden)
		return
	}

	mail := c.Param("mail")
	mail = strings.ToLower(mail)

	if !strings.HasSuffix(mail, "@cronos.de") {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{
				"the mail address must end with @cronos.de",
			},
		})
		return
	}

	ur, err := authClient.GetUserByEmail(context.Background(), mail)
	if err != nil {
		logrus.Info(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{
				"there was an error fetching the user from firebase",
			},
		})
		return
	}
	f := bson.D{{"user", ur.UID}}
	cur, err := client.Database("office_checkin").Collection("bookings").Find(context.Background(), f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}

	defer cur.Close(context.Background())
	var b Booking
	mind := time.Now().Add(- time.Hour * 24 * 14)
	relevantData := make(map[string]string)
	for cur.Next(context.Background()) {
		err := cur.Decode(&b)
		if err != nil {
			logrus.Warn(err)
			continue
		}
		t,err := time.Parse("2006-01-02", b.Date)
		if err != nil {
			logrus.Warn(err)
			continue
		}
		if mind.After(t) || t.After(time.Now()) {
			continue
		}
		relevantData[b.Date] = b.Area
	}

	contactData := []BacktracingItem{}
	for date, area := range relevantData {
		logrus.WithFields(logrus.Fields{
			"date": date,
			"area": area,
		}).Info("fetching data for backtracing")
		f := bson.D{{"area", area}, {"date", date}}
		cur, err := client.Database("office_checkin").Collection("bookings").Find(context.Background(), f)
		if err != nil {
			logrus.Error(err)
			continue
		}

		defer cur.Close(context.Background())
		for cur.Next(context.Background()) {
			_ = cur.Decode(&b)
			if b.UserName == mail {
				// Do not store own contact
				continue
			}
			contactData = append(contactData, BacktracingItem{
				Date:  date,
				Email: b.UserName,
				AreaName: b.AreaData.Name,
			})
		}
	}

	c.JSON(http.StatusOK, BacktracingResponse{
		For:       mail,
		CreatedAt: time.Now().String(),
		NotBefore: mind.String(),
		NotAfter: time.Now().String(),
		Data:      contactData,
	})
}