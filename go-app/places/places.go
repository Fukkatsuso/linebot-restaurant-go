package places

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Places is slice of place
type Places []Place

// APIResponse is format of places-API's response
type APIResponse struct {
	HTMLAttributions []interface{} `json:"html_attributions"`
	Places           Places        `json:"results"`
	Status           string        `json:"status"`
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
func Search(uri string) (Places, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var apiResp APIResponse
	json.Unmarshal(body, &apiResp)

	return apiResp.Places, nil
}
