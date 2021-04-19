package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signintech/gopdf"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

func handlePrintRequest(c *gin.Context) {
	if !c.GetBool("isAdmin") {
		c.AbortWithStatusJSON(http.StatusForbidden, ErrorForbidden)
		return
	}

	vb := getVisitorBookingsForDate(c.Param("date"))
	if len(vb) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	pdf := renderPDF(vb)
	c.Header("Content-Type", "application/pdf")
	c.Writer.Write(pdf.GetBytesPdf())

}

// renderPDF creates the actual pdf document, which is send to the user to print visitor cards.
// this function needs accomodations to match e.g. CI/CD
func renderPDF(vb []Visit) gopdf.GoPdf {
	doc := gopdf.GoPdf{}
	doc.Start(gopdf.Config{ PageSize: *gopdf.PageSizeA4 })
	doc.AddPage()
	err := doc.AddTTFFont("Roboto", "./assets/Roboto-Regular.ttf")
	if err != nil {
		logrus.Error(err)
	}

	doc.SetTextColor(getForegroundColor())

	offsetY := 25

	for i,v := range vb {
		offsetX := 8
		if i%2 == 1 {
			offsetX = 300
		}
		if i%2 == 0 && i != 0 {
			offsetY += 175
		}
		doc.SetX(float64(offsetX))
		doc.SetY(float64(offsetY))
		if err := doc.SetFont("Roboto", "", 36); err != nil {
			continue
		}
		if err := doc.Cell(nil, v.Visitor.LastName); err != nil {
			continue
		}
		doc.Br(40)
		doc.SetX(float64(offsetX))
		if err := doc.SetFont("Roboto", "", 24); err != nil {
			continue
		}
		if err := doc.Cell(nil, v.Visitor.FirstName); err != nil {
			continue
		}
		doc.Br(40)
		doc.SetX(float64(offsetX))
		if err := doc.SetFont("Roboto", "", 12); err != nil {
			continue
		}
		doc.Line(float64(offsetX),float64(offsetY + 87), float64(offsetX + 250), float64(offsetY + 87) )
		doc.Text(v.Visitor.Company)
		doc.Br(24)
		doc.SetX(float64(offsetX))
		doc.Text("Supervisor: " + v.Supervisor.DisplayName)
		doc.Br(18)
		doc.SetX(float64(offsetX))
		t,_ := time.Parse("2006-01-02", v.Date)
		doc.Text("Gültig am: " + t.Format("02.01.2006"))
		doc.SetFont("Roboto", "", 8)
		doc.Br(12)
		doc.SetX(float64(offsetX))

		doc.Text("Dieses Badge muss jederzeit gut sichtbar getragen werden.")
	}

	doc.SetLineType("dashed")
	doc.Line(290, 0, 290, 1500)
	doc.Line(0, 175, 1500, 175)
	doc.Line(0, 350, 1500, 350)
	doc.Line(0, 525, 1500, 525)
	doc.Line(0, 700, 1500, 700)
	return doc
}

func getForegroundColor() (uint8, uint8, uint8) {
	cc := cfg.Badge.ForegroundColor
	if len(cc) != 7 {
		logrus.Error("could not parse foreground color for badge. going to use default color black")
		return 0x00, 0x00, 0x00
	}
	if cc[0] != '#' {
		logrus.Error("could not parse foreground color for badge. going to use default color black")
		return 0x00, 0x00, 0x00
	}
	r, _ := strconv.Atoi(cc[1:2])
	g, _ := strconv.Atoi(cc[3:4])
	b, _ := strconv.Atoi(cc[5:6])
	return uint8(r), uint8(g), uint8(b)
}