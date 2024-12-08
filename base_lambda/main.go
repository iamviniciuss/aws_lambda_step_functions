package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	url := "https://www.ryanair.com/api/farfnd/v4/roundTripFares?departureAirportIataCode=MLA&outboundDepartureDateFrom=2024-12-01&market=en-us&adultPaxCount=1&outboundDepartureDateTo=2025-01-29&inboundDepartureDateFrom=2024-12-03&inboundDepartureDateTo=2025-01-31&durationFrom=2&durationTo=7&outboundDepartureDaysOfWeek=MONDAY,TUESDAY,WEDNESDAY,THURSDAY,FRIDAY,SATURDAY,SUNDAY&outboundDepartureTimeFrom=00:00&outboundDepartureTimeTo=23:59&inboundDepartureTimeFrom=00:00&inboundDepartureTimeTo=23:59"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Erro ao fazer a requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler a resposta: %v", err)
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Erro ao decodificar a resposta: %v", err)
	}

	for _, fare := range response.Fares {
		log.Println("Price:", fare.Summary.Price.Value, " - Country Name:", fare.Outbound.ArrivalAirport.CountryName, " - City: ", fare.Outbound.ArrivalAirport.City.Name, " - Date: ", fare.Outbound.DepartureDate, " -> ", fare.Outbound.ArrivalDate)
	}

	log.Printf("Decoded Response: %+v", len(response.Fares))
	log.Println("HELLO WORLD + LAMBDA")

	lambda.Start(func(ctxBase context.Context, sqsEvent *events.SQSEvent) error {
		log.Println("SQS EVENT OK")
		return nil
	})
}

type City struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	CountryCode string `json:"countryCode"`
}

type Airport struct {
	CountryName string `json:"countryName"`
	IataCode    string `json:"iataCode"`
	Name        string `json:"name"`
	SeoName     string `json:"seoName"`
	City        City   `json:"city"`
}

type Price struct {
	Value               float64 `json:"value"`
	ValueMainUnit       string  `json:"valueMainUnit"`
	ValueFractionalUnit string  `json:"valueFractionalUnit"`
	CurrencyCode        string  `json:"currencyCode"`
	CurrencySymbol      string  `json:"currencySymbol"`
}

type Flight struct {
	DepartureAirport Airport  `json:"departureAirport"`
	ArrivalAirport   Airport  `json:"arrivalAirport"`
	DepartureDate    string   `json:"departureDate"`
	ArrivalDate      string   `json:"arrivalDate"`
	Price            Price    `json:"price"`
	FlightKey        string   `json:"flightKey"`
	FlightNumber     string   `json:"flightNumber"`
	PreviousPrice    *float64 `json:"previousPrice"`
	PriceUpdated     int64    `json:"priceUpdated"`
}

type Summary struct {
	Price            Price    `json:"price"`
	PreviousPrice    *float64 `json:"previousPrice"`
	NewRoute         bool     `json:"newRoute"`
	TripDurationDays int      `json:"tripDurationDays"`
}

type Fare struct {
	Outbound Flight  `json:"outbound"`
	Inbound  Flight  `json:"inbound"`
	Summary  Summary `json:"summary"`
}

type Response struct {
	ArrivalAirportCategories interface{} `json:"arrivalAirportCategories"`
	Fares                    []Fare      `json:"fares"`
}
