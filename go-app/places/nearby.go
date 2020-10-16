package places

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

var placesAPIKey string

func init() {
	placesAPIKey = os.Getenv("GCP_PLACES_API_KEY")
}

// NearbyPlaces is a response of nearby-search
type NearbyPlaces struct {
	HTMLAttributions []interface{} `json:"html_attributions"`
	Results          []NearbyPlace `json:"results"`
	Status           string        `json:"status"`
}

// NearbyPlace is a part of format of API response
type NearbyPlace struct {
	Geometry struct {
		Location struct {
			Lat string `json:"lat"`
			Lng string `json:"lng"`
		} `json:"location"`
		Viewport struct {
			Northeast struct {
				Lat string `json:"lat"`
				Lng string `json:"lng"`
			} `json:"northeast"`
			Southwest struct {
				Lat string `json:"lat"`
				Lng string `json:"lng"`
			} `json:"southwest"`
		} `json:"viewport"`
	} `json:"geometry"`
	Icon         string `json:"icon"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	OpeningHours struct {
		OpenNow bool `json:"open_now"`
	} `json:"opening_hours,omitempty"`
	Photos []*struct {
		Height           int      `json:"height"`
		HTMLAttributions []string `json:"html_attributions"`
		PhotoReference   string   `json:"photo_reference"`
		Width            int      `json:"width"`
	} `json:"photos"`
	PlaceID  string `json:"place_id"`
	PlusCode struct {
		CompoundCode string `json:"compound_code"`
		GlobalCode   string `json:"global_code"`
	} `json:"plus_code"`
	PriceLevel       int      `json:"price_level,omitempty"`
	Rating           float64  `json:"rating"`
	Reference        string   `json:"reference"`
	Scope            string   `json:"scope"`
	Types            []string `json:"types"`
	UserRatingsTotal int      `json:"user_ratings_total"`
	Vicinity         string   `json:"vicinity"`
}

// MarshalPlace converts NearbyPlace to Place
func (p *NearbyPlace) MarshalPlace() Place {
	params := map[string]string{
		"key":      placesAPIKey,
		"maxwidth": "350",
	}
	return Place{
		PlaceID:      p.PlaceID,
		Name:         p.Name,
		Rating:       p.Rating,
		PhotoURI:     p.PhotoURI(params),
		GooglemapURI: p.GooglemapURI(),
	}
}

// MarshalPlaces converts NearbyPlaces to Places
func (p *NearbyPlaces) MarshalPlaces() Places {
	places := make(Places, 0)
	for i := range p.Results {
		places = append(places, p.Results[i].MarshalPlace())
	}
	return places
}

// PhotoURI returns uri
func (p *NearbyPlace) PhotoURI(params map[string]string) string {
	if len(p.Photos) == 0 {
		return AlternativePhotoURI()
	}
	params["photoreference"] = p.Photos[0].PhotoReference
	return GooglemapPhotoURI(params)
}

// AlternativePhotoURI returns uri of line-cdn-clip
func AlternativePhotoURI() string {
	id := "12" // 1 ~ 13
	clip := "clip" + id + ".jpg"
	uri := "https://scdn.line-apps.com/n/channel_devcenter/img/flexsnapshot/clip/" + clip
	return uri
}

// ErrRedirectAttempted errors
var ErrRedirectAttempted = errors.New("redirect")

// GooglemapPhotoURI returns uri of googlemap-photo
func GooglemapPhotoURI(params map[string]string) string {
	uri := "https://maps.googleapis.com/maps/api/place/photo?"
	for k, v := range params {
		if uri[len(uri)-1] != '?' {
			uri += "&"
		}
		uri += k + "=" + v
	}

	client := &http.Client{
		Timeout: time.Duration(3) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return ErrRedirectAttempted
		},
	}
	resp, err := client.Head(uri)
	if urlError, ok := err.(*url.Error); !(ok && urlError.Err == ErrRedirectAttempted) {
		return AlternativePhotoURI()
	}
	defer resp.Body.Close()
	return resp.Header["Location"][0]
}

// GooglemapURI returns uri of the place on googlemap
func (p *NearbyPlace) GooglemapURI() string {
	uri := "https://www.google.com/maps/search/?api=1"
	uri += "&query=" + p.Geometry.Location.Lat + "," + p.Geometry.Location.Lng
	uri += "&query_place_id=" + p.PlaceID
	return uri
}
