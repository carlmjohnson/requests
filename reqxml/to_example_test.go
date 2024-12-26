package reqxml_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/carlmjohnson/requests/reqxml"
)

func init() {
	http.DefaultClient.Transport = reqtest.Recorder(reqtest.ModeReplay, nil, "testdata")
}

func ExampleTo() {
	type CD struct {
		Title   string  `xml:"TITLE"`
		Artist  string  `xml:"ARTIST"`
		Country string  `xml:"COUNTRY"`
		Price   float64 `xml:"PRICE"`
		Year    int     `xml:"YEAR"`
	}
	type Catalog struct {
		CDs []CD `xml:"CD"`
	}
	var cat Catalog
	err := requests.
		URL("https://www.w3schools.com/xml/cd_catalog.xml").
		Handle(reqxml.To(&cat)).
		Fetch(context.Background())
	if err != nil {
		fmt.Println("Error!", err)
	}
	for _, cd := range cat.CDs {
		if cd.Price > 10 && cd.Year < 1990 {
			fmt.Printf("%s - %s $%.2f", cd.Artist, cd.Title, cd.Price)
		}
	}
	// Output:
	// Bob Dylan - Empire Burlesque $10.90
}
