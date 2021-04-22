package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func addBooking(c *gin.Context) {
	userId, _ := c.Get("userId")
	var br AddBookingRequest
	if err := c.BindJSON(&br); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"body malformed. could not parse JSON"},
		})
		return
	}

	dateLayout := "2006-01-02"

	if br.End != "" && br.Start == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code: http.StatusBadRequest,
			Errors: []string{
				"you have to provide a start date if providing an end date",
			},
		})
		return
	}

	if br.End == "" && br.Start != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code: http.StatusBadRequest,
			Errors: []string{
				"you have to provide an end date if providing a start date",
			},
		})
		return
	}

	if br.End != "" && br.Start != "" {
		genDates := []string{}
		errs := []string{}
		start, err := time.Parse(dateLayout, br.Start)
		if err != nil {
			errs = append(errs, "could not parse start date")
		}
		end, err := time.Parse(dateLayout, br.End)
		if err != nil {
			errs = append(errs, "could not parse end date")
			goto parsingErrorChecking
		}
		if end.Before(start) {
			errs = append(errs, "end is before start date")
		}
		if end.Equal(start) {
			errs = append(errs, "end is equal to start date")
		}
	parsingErrorChecking:
		if len(errs) > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
				Code:   http.StatusBadRequest,
				Errors: errs,
			})
			return
		}
		tdiff := int(end.Sub(start).Hours() / 24)
		iwq := c.Query("include-weekend")
		iw := iwq == "yes" || iwq == "true" || iwq == "1"
		for i := 0; i <= tdiff; i++ {
			t := start.Add(time.Duration(i) * time.Hour * 24)
			// Do not store weekend dates
			if (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) && !iw {
				continue
			}
			genDates = append(genDates, t.Format(dateLayout))
		}

		// override specified dates and only use generated dates in range
		br.Dates = genDates
	}

	if len(br.Dates) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"you must provide at least 1 date"},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()
	for _, d := range br.Dates {
		var booking Booking
		booking.Area = br.Area
		booking.Date = d
		t, err := time.Parse(dateLayout, booking.Date)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
				Code:   http.StatusBadRequest,
				Errors: []string{"could not parse date"},
			})
			return
		}
		ct := time.Now()
		ts := time.Date(ct.Year(), ct.Month(), ct.Day(), 0, 0, 0, 0, time.UTC)
		if t.Before(ts) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
				Code:   http.StatusBadRequest,
				Errors: []string{"date is in the past. you have to book a date in the future or today"},
			})
			return
		}
		if !isAreaBookableForDate(booking.Area, booking.Date) {
			c.AbortWithStatusJSON(http.StatusLocked, ErrorResponse{
				Code:   http.StatusLocked,
				Errors: []string{"no capacity for booking on your specified date"},
			})
			return
		}
		booking.User = fmt.Sprintf("%v", userId)
		filter := bson.D{{"user", userId}, {"date", booking.Date}}
		var existingBooking Booking
		err = client.Collection("bookings").FindOne(ctx, filter).Decode(&existingBooking)
		if err != nil && err != mongo.ErrNoDocuments {
			logrus.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
			return
		}
		if err != nil && err == mongo.ErrNoDocuments {
			un, _ := c.Get("userMail")
			booking.UserName = fmt.Sprintf("%v", un)
			aid, _ := primitive.ObjectIDFromHex(booking.Area)
			err = client.Collection("areas").FindOne(ctx, bson.D{{"_id", aid}}).Decode(&booking.AreaData)
			_, err = client.Collection("bookings").InsertOne(ctx, booking)
			if err != nil {
				logrus.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
				return
			}

		}
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:   http.StatusConflict,
				Errors: []string{"you already checked in for that date"},
			})
		}

	}

	c.JSON(http.StatusOK, nil)
}

func getBookings(c *gin.Context) {
	ufc, _ := c.Get("userId")
	uid := fmt.Sprintf("%v", ufc)
	filter := bson.D{{"user", uid}}
	cur, err := client.Collection("bookings").Find(context.Background(), filter, nil)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	var bookings []Booking
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		booking := Booking{}
		err := cur.Decode(&booking)
		if err != nil {
			logrus.Error(err)
		}
		booking.ID = cur.Current.Lookup("_id").ObjectID().Hex()
		aid, _ := primitive.ObjectIDFromHex(booking.Area)
		err = client.Collection("areas").FindOne(context.Background(), bson.D{{"_id", aid}}).Decode(&booking.AreaData)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
			return
		}
		bookings = append(bookings, booking)
	}
	c.JSON(http.StatusOK, Bookings{Bookings: bookings})
}

func updateBooking(c *gin.Context) {
	logrus.Debug("Updating single booking")
	c.JSON(http.StatusNotImplemented, ErrorNotImplemented)
}

func deleteBooking(c *gin.Context) {
	bid := c.Param("id")
	logrus.WithField("booking_id", bid).Trace("Going to delete single booking")
	uic, _ := c.Get("userId")
	uid := fmt.Sprintf("%v", uic)
	pbid, err := primitive.ObjectIDFromHex(bid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"object id is not in proper format"},
		})
		return
	}
	f := bson.D{{"_id", pbid}, {"user", uid}}
	r, err := client.Collection("bookings").DeleteOne(context.Background(), f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	if r.DeletedCount > 0 {
		c.JSON(http.StatusOK, SuccessResponse{
			Code:    http.StatusOK,
			Message: "successfully deleted booking",
		})
		return

	}
	c.JSON(http.StatusNoContent, SuccessResponse{
		Code:    http.StatusNoContent,
		Message: "no booking deleted",
	})
}

func adminGetBookings(c *gin.Context) {
	logrus.Debug("getting all bookings for admin dashboard")
	var bookings []Booking
	f := bson.D{}
	cursor, err := client.Collection("bookings").Find(context.Background(), f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
			Code:   http.StatusInternalServerError,
			Errors: []string{"internal server error"},
		})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		booking := Booking{}
		err := cursor.Decode(&booking)
		if err != nil {
			logrus.Error(err)
		}
		booking.ID = cursor.Current.Lookup("_id").ObjectID().Hex()
		aid, _ := primitive.ObjectIDFromHex(booking.Area)
		err = client.Collection("areas").FindOne(context.Background(), bson.D{{"_id", aid}}).Decode(&booking.AreaData)
		bookings = append(bookings, booking)
	}
	c.JSON(http.StatusOK, bookings)
}

func adminGetBookingsForDate(c *gin.Context) {
	logrus.Debug("getting all bookings for given date for admin dashboard")

	dateLayout := "2006-01-02"

	date := c.Param("date")
	_, err := time.Parse(dateLayout, date)
	if err != nil {
		logrus.Info(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"invalid date"},
		})
		return
	}
	f := bson.D{{"date", date}}
	cursor, err := client.Collection("bookings").Find(context.Background(), f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	var bookings []Booking
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		booking := Booking{}
		err := cursor.Decode(&booking)
		if err != nil {
			logrus.Error(err)
		}
		booking.ID = cursor.Current.Lookup("_id").ObjectID().Hex()
		aid, _ := primitive.ObjectIDFromHex(booking.Area)
		err = client.Collection("areas").FindOne(context.Background(), bson.D{{"_id", aid}}).Decode(&booking.AreaData)
		bookings = append(bookings, booking)
	}

	v := getVisitorBookingsForDate(date)

	c.JSON(http.StatusOK, struct {
		Visits   []Visit   `json:"visits"`
		Bookings []Booking `json:"bookings"`
	}{
		Visits:   v,
		Bookings: bookings,
	})
}
