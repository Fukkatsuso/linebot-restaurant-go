package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/bot"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/config"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
)

// QueryToMap converts Query to map[string]string
func QueryToMap(query *bot.Query) map[string]string {
	params := map[string]string{
		"key":      config.GCPPlacesAPIKey,
		"type":     "restaurant",
		"location": query.Lat + "," + query.Lng,
		"radius":   query.Radius,
	}
	if len(query.Keywords) > 0 {
		params["keyword"] = strings.Join(query.Keywords, "+")
	}
	return params
}

// BuildURI returns uri with parameters
func BuildURI(apiType string, params map[string]string) string {
	uri := "https://maps.googleapis.com/maps/api/place/" + apiType + "/json?language=ja"
	for k, v := range params {
		uri += "&" + k + "=" + v
	}
	return uri
}

// NearbySearch gets places
func NearbySearch(query *bot.Query, p *places.Places) (string, error) {
	uri := BuildURI("nearbysearch", QueryToMap(query))
	resp, err := http.Get(uri)
	if err != nil {
		return uri, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return uri, err
	}
	var nearby places.NearbyPlaces
	json.Unmarshal(body, &nearby)
	*p = nearby.MarshalPlaces()

	return uri, nil
}

// DetailsSearch gets details
func DetailsSearch(placeID string, p *places.Place) (string, error) {
	uri := BuildURI("details", map[string]string{"placeid": placeID, "key": config.GCPPlacesAPIKey})
	resp, err := http.Get(uri)
	if err != nil {
		return uri, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return uri, err
	}
	var details places.PlaceDetails
	json.Unmarshal(body, &details)
	*p = details.Result.MarshalPlace()

	return uri, nil
}
