package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strconv"
	"time"
)

func getAreas(c *gin.Context) {

	logrus.Debug("Fetching all areas")
	filter := bson.D{}
	cur, err := client.Collection("areas").Find(context.Background(), filter, nil)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	var areas []Area

	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		area := Area{}
		err := cur.Decode(&area)
		if err != nil {
			log.Fatal(err)
		}
		area.ID = cur.Current.Lookup("_id").ObjectID().Hex()
		areas = append(areas, area)
	}
	c.JSON(http.StatusOK, Areas{Areas: areas})
}

func getArea(c *gin.Context) {
	logrus.Debug("fetching single area")
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", objID}}

	area := Area{}
	err := client.Collection("areas").FindOne(context.Background(), filter).Decode(&area)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	area.ID = id

	c.JSON(http.StatusOK, area)
}

func unavailableDates(c *gin.Context) {
	a := c.Param("id")
	logrus.WithField("area", a).Debug("looking for fully booked dates for area")
	f := bson.D{{"area", a}}
	ad := getAreaFromDB(a)
	cur, err := client.Collection("bookings").Find(context.Background(), f)

	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	defer cur.Close(context.Background())

	dates := make(map[string]uint16, 256)
	var fullDates = []string{}

	for cur.Next(context.Background()) {
		b := Booking{}
		err := cur.Decode(&b)
		if err != nil {
			log.Fatal(err)
		}
		if _, ok := dates[b.Date]; ok {
			dates[b.Date]++
		} else {
			dates[b.Date] = 1
		}
		if dates[b.Date] == ad.Capacity {
			fullDates = append(fullDates, b.Date)
		}
	}

	c.JSON(http.StatusOK, struct {
		Dates []string `json:"dates"`
	}{Dates: fullDates})

}

func updateArea(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorNotImplemented)
}

func deleteArea(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorNotImplemented)
}

func getForecast(c *gin.Context) {
	today := time.Now()
	a := c.Param("id")
	qdif := c.Query("days-in-future")
	dif := 0
	if qdif == "" {
		dif = 7
	} else {
		var err error
		dif, err = strconv.Atoi(qdif)
		if err != nil {
			logrus.Info(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
				Code:   http.StatusBadRequest,
				Errors: []string{"cannot parse days-in-future query parameter"},
			})
			return
		}
	}

	if dif < 1 || dif > 28*4 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code: http.StatusBadRequest,
			Errors: []string{
				"days-in-future must be in range 1 to 112",
			},
		})
		return
	}

	ub := make(map[string]bool)
	uid := c.GetString("userId")
	filter := bson.D{{"user", uid}}
	cur, err := client.Collection("bookings").Find(context.Background(), filter, nil)
	if err != nil {
		logrus.Warn(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		var b Booking
		err := cur.Decode(&b)
		if err != nil {
			logrus.Warn(err)
			continue
		}
		ub[b.Date] = true
	}

	ad := getAreaFromDB(a)

	dateLayout := "2006-01-02"
	var i int32
	forecast := make(map[string]uint16, 10)

	for i = 0; len(forecast) < dif; {
		t := today.Add(24 * time.Hour * time.Duration(i))
		i++
		if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
			continue
		}
		date := t.Format(dateLayout)
		b := getBookingsForDate(a, date)
		forecast[date] = b
		logrus.WithFields(logrus.Fields{
			"bookings": b, "date": date, "area": a,
		}).Trace("forecast for date")
	}

	fis := []ForecastItem{}

	for key, value := range forecast {
		fis = append(fis, ForecastItem{
			Date:           key,
			BookedSeats:    value,
			BookedByMyself: ub[key],
		})
	}

	c.JSON(http.StatusOK, Forecast{
		CreatedAt: time.Now().String(),
		Bookings:  fis,
		Area:      ad,
	})

}
