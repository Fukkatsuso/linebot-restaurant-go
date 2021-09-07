package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
)

type SearchType string

const (
	SearchTypeNearby  SearchType = "nearbysearch"
	SearchTypeDetails SearchType = "details"
)

// NearbySearch
func (bot *Bot) NearbySearch(query *Query) (*places.Places, error) {
	uri := buildURI(SearchTypeNearby, bot.nearbySearchParams(query))
	fmt.Println("[URI]", uri)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var nearby places.NearbyPlaces
	json.Unmarshal(body, &nearby)

	p := nearby.MarshalPlaces(bot.GCPPlacesAPIKey)
	return &p, nil
}

// DetailsSearch
func (bot *Bot) DetailsSearch(placeID string) (*places.Place, error) {
	uri := buildURI(SearchTypeDetails, bot.detailsSearchMap(placeID))
	fmt.Println("[URI]", uri)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var details places.PlaceDetails
	json.Unmarshal(body, &details)

	p := details.Result.MarshalPlace()
	return &p, nil
}

// make nearby search params
func (bot *Bot) nearbySearchParams(query *Query) map[string]string {
	params := map[string]string{
		"key":      bot.GCPPlacesAPIKey,
		"type":     "restaurant",
		"location": query.Lat + "," + query.Lng,
		"radius":   query.Radius,
	}
	if len(query.Keywords) > 0 {
		params["keyword"] = strings.Join(query.Keywords, "+")
	}
	return params
}

// make details search params
func (bot *Bot) detailsSearchMap(placeID string) map[string]string {
	return map[string]string{
		"placeid": placeID,
		"key":     bot.GCPPlacesAPIKey,
	}
}

// biuld uri with params
func buildURI(searchType SearchType, params map[string]string) string {
	endpoint := "https://maps.googleapis.com/maps/api/place/"

	uri := endpoint + string(searchType) + "/json?language=ja"
	for k, v := range params {
		uri += fmt.Sprintf("&%s=%s", k, v)
	}

	return uri
}
