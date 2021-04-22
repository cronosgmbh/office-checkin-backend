package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func getBookingsForDate(area, date string) uint16 {
	if area == "" || date == "" {
		return 0
	}
	filter := bson.D{{"date", date}, {"area", area}}
	res, err := client.Collection("bookings").CountDocuments(context.Background(), filter, nil)
	if err != nil {
		logrus.Error(err)
		return 0
	}
	return uint16(res)
}

func getAreaFromDB(area string) (a Area) {
	objID, _ := primitive.ObjectIDFromHex(area)
	filter := bson.D{{"_id", objID}}
	_ = client.Collection("areas").FindOne(context.Background(), filter).Decode(&a)
	a.ID = area
	return
}

func isAreaBookableForDate(area, date string) bool {
	return getBookingsForDate(area, date) < getAreaFromDB(area).Capacity
}

func getVisitorBookingsForDate(date string) []Visit {
	f := bson.D{{"date", date}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cur, err := client.Collection("visits").Find(ctx, f)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	defer cur.Close(ctx)
	visits := []Visit{}
	v := Visit{}
	for cur.Next(ctx) {
		if err := cur.Decode(&v); err != nil {
			logrus.Error(err)
			continue
		}
		visits = append(visits, v)
	}

	return visits
}

func isDateBookableForVisitor(date string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	f := bson.D{{"date", date}}
	c, err := client.Collection("visits").CountDocuments(ctx, f)
	if err != nil {
		logrus.Error(err)
		return false
	}
	return c < 400 // Actually reduce number of bookings allowed per day, but for now let it stick at 400
}
