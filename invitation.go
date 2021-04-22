package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"time"
)

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown from server")
		}
	}
	return nil, nil
}

func acceptInvitation(c *gin.Context) {
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
	update := bson.M{
		"$set": struct {
			HasAccepted bool
		}{HasAccepted: true},
	}
	if _, err := client.Collection("visits").UpdateOne(ctx, f, update); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorInternalError)
		return
	}
	c.Status(http.StatusOK)
}

func resendMail(c *gin.Context) {

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

	go func() {
		date, _ := time.Parse("2006-01-02", visit.Date)
		if err := sendMail(visit.Visitor.Email, visit.ID.Hex(), visit.Visitor.FirstName+" "+visit.Visitor.LastName, date.Format("02.01.2006")); err != nil {
			logrus.Error(err)
		}
	}()
	c.Status(http.StatusOK)
}

func sendMail(address, visitID, name, date string) error {
	logrus.Debug("trying to send mail")
	from := mail.Address{cfg.Email.FromName, cfg.Email.FromMail}
	password := cfg.Email.Password
	to := mail.Address{name, address}
	smtpHost := fmt.Sprintf("%s:%s", cfg.Email.Host, cfg.Email.Port)
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\nFrom: " + cfg.Email.FromName + " <" + cfg.Email.FromMail + ">\n\n"

	var body bytes.Buffer

	body.Write([]byte(fmt.Sprintf("Subject: Anmeldung als Gast bei der cronos Unternehmensberatung \n%s\n\n", mimeHeaders)))

	t, err := template.ParseFiles("mail-templates/invitation.html")
	if err != nil {
		return err
	}
	err = t.Execute(&body, struct {
		VisitID string
		Date    string
		Name    string
	}{
		VisitID: visitID,
		Date:    date,
		Name:    name,
	})
	if err != nil {
		return err
	}
	host, _, _ := net.SplitHostPort(smtpHost)

	tlsconfig := &tls.Config{
		ServerName: host,
	}

	conn, err := net.Dial("tcp", smtpHost)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err := c.StartTLS(tlsconfig); err != nil {
		return err
	}

	if err = c.Auth(LoginAuth(from.Address, password)); err != nil {
		return err
	}

	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(body.Bytes())
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
