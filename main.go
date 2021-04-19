package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

var client *mongo.Client
var cfg Config

func main() {

	loadConfig(&cfg)

	switch cfg.Service.Environment {
	case "development":
		logrus.SetLevel(logrus.TraceLevel)
	case "staging":
		logrus.SetLevel(logrus.DebugLevel)
	case "production":
		logrus.SetLevel(logrus.WarnLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithFields(logrus.Fields{"cronos_env": cfg.Service.Environment, "port": cfg.Service.Port}).Info("Starting office checkin backend service")

	initFirebase()
	client = connectToDB()
	initSettings()

	gin.SetMode(gin.ReleaseMode)
	e := gin.New()

	api := e.Group("v1")
	api.Use(corsHeader())
	api.Use(cors.Default())

	areas := api.Group("areas")
	areas.Use(cors.Default(), authMiddleware())
	areas.OPTIONS("")
	areas.OPTIONS(":id")

	areas.GET("", getAreas)
	areas.DELETE("", deleteArea)
	areas.GET(":id", getArea)
	areas.PATCH(":id", updateArea)

	areas.GET(":id/unavailable-dates", unavailableDates)
	areas.OPTIONS(":id/unavailable-dates")

	areas.GET(":id/forecast", getForecast)
	areas.OPTIONS(":id/forecast")

	bookings := api.Group("bookings")
	bookings.Use(cors.Default(), authMiddleware())
	bookings.OPTIONS("")
	bookings.OPTIONS(":id")

	bookings.GET("", getBookings)
	bookings.POST("", addBooking)
	bookings.DELETE(":id", deleteBooking)
	bookings.PATCH(":id", updateBooking)

	admin := api.Group("admin")
	admin.Use(cors.Default(), authMiddleware())

	admin.GET("bookings", adminGetBookings)
	admin.GET("bookings/:date", adminGetBookingsForDate)
	admin.GET("refresh-settings", refreshSettingsHandler)
	admin.GET("users/:mail/covid-backtracing", covidBacktracing)
	admin.GET("visitor-badges/:date", handlePrintRequest)

	admin.OPTIONS("bookings")
	admin.OPTIONS("bookings/:date")
	admin.OPTIONS("refresh-settings")
	admin.OPTIONS("users/:mail/covid-backtracing")
	admin.OPTIONS("visitor-badges/:date")

	user := api.Group("user")
	user.Use(cors.Default(), authMiddleware())
	user.OPTIONS("")
	user.GET("", getUser)
	user.PUT("", updateUser)


	users := api.Group("users")
	users.Use(cors.Default(), authMiddleware())
	users.OPTIONS("")

	users.POST("is-admin", customClaims)
	users.OPTIONS("is-admin")

	visitors := api.Group("visitors")
	visitors.Use(cors.Default(), authMiddleware())
	visitors.POST("", addVisitor)
	visitors.GET("", getVisitors)
	visitors.GET(":id", getVisitor)
	visitors.DELETE(":id", deleteVisitor)
	visitors.OPTIONS("")
	visitors.OPTIONS(":id")

	visits := api.Group("visits")
	visits.Use(cors.Default(), authMiddleware())
	visits.POST("", addVisit)
	visits.GET("", getVisits)
	visits.GET(":id")
	visits.DELETE(":id", deleteVisit)
	visits.OPTIONS("")
	visits.OPTIONS(":id")

	invitations := api.Group("invitations")
	invitations.Use(cors.Default())
	invitations.GET(":id", getSingleVisit)
	invitations.PATCH(":id", acceptInvitation)
	invitations.OPTIONS(":id")
	invitations.POST(":id/resend-mail", authMiddleware(), resendMail)
	invitations.OPTIONS(":id/resend-mail")

	go runTasks()

	logrus.Fatal(http.ListenAndServe(":3000", e))

}
