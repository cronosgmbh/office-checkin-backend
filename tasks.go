package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func runTasks() {
	deleteOldBookings()
	interval, err := time.ParseDuration(cfg.Service.TaskInterval)
	if err != nil {
		logrus.Warnf("could not parse duration %s. going to use default 15 minute interval time for tasks", cfg.Service.TaskInterval)
		interval = time.Minute * 15
	}
	for range time.Tick(interval) {
		go deleteOldBookings()
	}
}

func deleteOldBookings() {
	if !cfg.Bookings.AutoDelete {
		return
	}
	ds := fmt.Sprintf("%dh", cfg.Bookings.DeleteAfterDays * 24)
	age, err := time.ParseDuration(ds)
	if err != nil {
		age = time.Hour * 24 * 14
		logrus.Warnf("could not parse duration. going to use default age for deleting old bookings")
	}
	d := time.Now().Add(- age).Format("2006-01-02")
	logrus.WithField("before_date", d).Info("executing delete old bookings task")
	filter := bson.M{
		"date": bson.M{"$lte": d},
	}
	dr, err := client.Database("office_checkin").Collection("bookings").DeleteMany(context.Background(), filter)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.WithField("deleted_items", dr.DeletedCount).Info("executed delete old bookings task")
}