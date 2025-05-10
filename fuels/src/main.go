package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ResponseBody struct {
	Success bool     `json:"success"`
	Center  Location `json:"center"`
	Results []Result `json:"results"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type RequestBody struct {
	Points []Point `json:"points"`
	Radius int     `json:"radius"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Fuel struct {
	ID     int     `json:"id"`
	Price  float64 `json:"price"`
	Name   string  `json:"name"`
	FuelID int     `json:"fuel_id"`
	IsSelf bool    `json:"is_self"`
}

type Result struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Fuels      []Fuel    `json:"fuels"`
	Location   Location  `json:"location"`
	InsertDate time.Time `json:"insertDate"`
	Address    *string   `json:"address"`
	Brand      string    `json:"brand"`
	Distance   string    `json:"distance"`
}

func Scrape(fuelType string, location *Location, radius int) ResponseBody {
	data := map[string]interface{}{
		"points": []map[string]float64{
			{"lat": location.Lat, "lng": location.Lon},
		},
		"fuelType": fuelType,
		"radius":   radius,
	}

	requestData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	url := "https://carburanti.mise.gov.it/ospzApi/search/zone"

	postRequest, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		fmt.Println(err)
	}

	postRequest.Header.Add("Content-Type", "application/json")
	postRequest.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:137.0) Gecko/20100101 Firefox/137.0")

	client := &http.Client{}
	response, err := client.Do(postRequest)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseData ResponseBody
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.Fatal(err)
	}

	return responseData
}

func main() {

	location := Location{
		Lat: 45.6893093,
		Lon: 8.735909099999999,
	}
	fuelType := "1-x"
	responseData := Scrape(fuelType, &location, 5)
	fmt.Println("Success: ", responseData.Success)
	fmt.Printf("Center: Lat %f, Lon %f\n", responseData.Center.Lat, responseData.Center.Lon)
	fmt.Printf("Number of results: %d\n", len(responseData.Results))
	for _, result := range responseData.Results {
		fmt.Println(result)
	}

}
