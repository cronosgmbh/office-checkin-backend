package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type Visitor struct {
	ID        primitive.ObjectID `bson:"_id" ,json:"ID"`
	FirstName string             `bson:"first_name" ,json:"FirstName"`
	LastName  string             `bson:"last_name" ,json:"LastName"`
	Email     string             `bson:"email" ,json:"Email"`
	Phone     string             `bson:"phone" ,json:"Phone"`
	Company   string             `bson:"company" ,json:"Company"`
	CreatedBy string             `bson:"created_by" ,json:"CreatedBy"`
}

type Supervisor struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type Visit struct {
	ID                primitive.ObjectID `bson:"_id"`
	Visitor           Visitor            `json:"visitor"`
	Date              string             `json:"date"`
	AdditionalInfo    string             `json:"additional_info"`
	NeedsParkingSpace bool               `json:"needs_parking_space"`
	User              string             `json:"user"`
	Supervisor        Supervisor         `json:"supervisor"`
	HasAccepted       bool               `json:"has_accepted"`
}

func addVisitor(c *gin.Context) {
	var v Visitor
	err := c.BindJSON(&v)
	if err != nil {
		logrus.Debug(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code: http.StatusBadRequest,
			Errors: []string{
				"body malformed",
			},
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	_, err = client.Collection("visitors").InsertOne(ctx, v)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	c.Status(http.StatusOK)
}

func getVisitor(c *gin.Context) {
	id := c.Param("id")
	var v Visitor
	pid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logrus.Info(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"cannot parse id"},
		})
		return
	}
	f := bson.D{{"_id", pid}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()
	err = client.Collection("visitors").FindOne(ctx, f).Decode(&v)
	if err != nil {
		logrus.Error(err)
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
			return
		}
	}

	c.JSON(http.StatusOK, v)
}

func getVisitors(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	f := bson.D{}
	rows, err := client.Collection("visits").Find(ctx, f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	v := make(map[string]Visitor)
	tv := Visit{}
	for rows.Next(ctx) {
		err := rows.Decode(&tv)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
			return
		}
		if _, ok := v[tv.Visitor.Email]; !ok {
			v[tv.Visitor.Email] = tv.Visitor
		}
	}

	c.JSON(200, v)
}

func deleteVisit(c *gin.Context) {
	uid := c.GetString("userId")
	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		logrus.Info(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"visit id malformed"},
		})
		return
	}
	var f bson.D
	if c.GetBool("isAdmin") {
		f = bson.D{{"_id", oid}}
	} else {
		f = bson.D{{"_id", oid}, {"user", uid}}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	dr, err := client.Collection("visits").DeleteOne(ctx, f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	c.JSON(http.StatusOK, struct {
		DeletedItems int64 `json:"deleted_items"`
	}{
		DeletedItems: dr.DeletedCount,
	})
}

func deleteVisitor(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorNotImplemented)
}

func addVisit(c *gin.Context) {
	r := Visit{}
	if err := c.BindJSON(&r); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"body malformed"},
		})
		return
	}
	_, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"date malformed. must be yyyy-mm-dd"},
		})
		return
	}
	if !isDateBookableForVisitor(r.Date) {
		c.AbortWithStatusJSON(http.StatusConflict, ErrorResponse{
			Code:   http.StatusConflict,
			Errors: []string{"there are more than 4 bookings for a visitor on this date. you have to select another date"},
		})
		return
	}
	u, err := getSingleUser(c.GetString("userId"))
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	r.Supervisor = Supervisor{
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DisplayName: u.FirstName + " " + u.LastName,
		Email:       u.Email,
	}

	r.User = c.GetString("userId")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	r.ID = primitive.NewObjectID()
	r.Visitor.ID = primitive.NewObjectID()
	r.HasAccepted = false
	_, err = client.Collection("visits").InsertOne(ctx, r)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}

	go func() {
		date, _ := time.Parse("2006-01-02", r.Date)
		if err = sendMail(r.Visitor.Email, r.ID.Hex(), r.Visitor.FirstName+" "+r.Visitor.LastName, date.Format("02.01.2006")); err != nil {
			logrus.Error(err)
		}
	}()
	c.JSON(http.StatusOK, r)

}

func getVisits(c *gin.Context) {
	uid := c.GetString("userId")
	f := bson.D{{"user", uid}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cur, err := client.Collection("visits").Find(ctx, f)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorInternalError)
		return
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
	c.JSON(http.StatusOK, visits)
}

func getSingleVisit(c *gin.Context) {
	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		logrus.Info(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:   http.StatusBadRequest,
			Errors: []string{"visit id malformed"},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	f := bson.D{{"_id", oid}}

	visit := Visit{}

	err = client.Collection("visits").FindOne(ctx, f).Decode(&visit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:   http.StatusNotFound,
				Errors: []string{"the visit could not be found"},
			})
			return
		}
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}

	// Alwasy reset these information, because they are just used internally to provide more info to coworkers
	// but should not be seen by any guest
	visit.AdditionalInfo = ""
	visit.NeedsParkingSpace = false
	visit.User = ""
	c.JSON(http.StatusOK, visit)
}
