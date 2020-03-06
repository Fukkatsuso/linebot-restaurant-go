package places

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// APIResponse is format of places-API's response
type APIResponse struct {
	HTMLAttributions []interface{} `json:"html_attributions"`
	Results          []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			Viewport struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		Icon         string `json:"icon"`
		ID           string `json:"id"`
		Name         string `json:"name"`
		OpeningHours struct {
			OpenNow bool `json:"open_now"`
		} `json:"opening_hours,omitempty"`
		PlaceID  string `json:"place_id"`
		PlusCode struct {
			CompoundCode string `json:"compound_code"`
			GlobalCode   string `json:"global_code"`
		} `json:"plus_code"`
		Rating           int      `json:"rating"`
		Reference        string   `json:"reference"`
		Scope            string   `json:"scope"`
		Types            []string `json:"types"`
		UserRatingsTotal int      `json:"user_ratings_total"`
		Vicinity         string   `json:"vicinity"`
		Photos           []struct {
			Height           int      `json:"height"`
			HTMLAttributions []string `json:"html_attributions"`
			PhotoReference   string   `json:"photo_reference"`
			Width            int      `json:"width"`
		} `json:"photos,omitempty"`
		PriceLevel int `json:"price_level,omitempty"`
	} `json:"results"`
	Status string `json:"status"`
}

// BuildURI returns uri with parameters
func BuildURI(apiType string, params map[string]string) string {
	uri := "https://maps.googleapis.com/maps/api/place/" + apiType + "/json?"
	for k, v := range params {
		if uri[len(uri)-1] != '?' {
			uri += "&"
		}
		uri += k + "=" + v
	}
	return uri
}

// Search gets places and returns them by struct format
func Search(uri string) (*APIResponse, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var places APIResponse
	json.Unmarshal(body, &places)

	return &places, nil
}
