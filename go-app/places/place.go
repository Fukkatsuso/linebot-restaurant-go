package places

import (
	"strconv"
)

// Place is a part of the format of places-API's response
type Place struct {
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
	Photos           []*struct {
		Height           int      `json:"height"`
		HTMLAttributions []string `json:"html_attributions"`
		PhotoReference   string   `json:"photo_reference"`
		Width            int      `json:"width"`
	} `json:"photos"`
	PriceLevel int `json:"price_level,omitempty"`
}

// Latitude returns the latitude of the place
func (place *Place) Latitude() string {
	return strconv.FormatFloat(place.Geometry.Location.Lat, 'f', -1, 64)
}

// Longitude returns the longitude of the place
func (place *Place) Longitude() string {
	return strconv.FormatFloat(place.Geometry.Location.Lng, 'f', -1, 64)
}

// PhotoURI returns uri
func (place *Place) PhotoURI(params map[string]string) string {
	if len(place.Photos) == 0 {
		return AlternativePhotoURI()
	}
	params["photoreference"] = place.Photos[0].PhotoReference
	return GoogleMapPhotoURI(params)
}

// GoogleMapURI returns uri of the place on googlemap
func (place *Place) GoogleMapURI() string {
	uri := "https://www.google.com/maps/search/?api=1"
	uri += "&query=" + place.Latitude() + "," + place.Longitude()
	uri += "&query_place_id" + place.ID
	return uri
}
