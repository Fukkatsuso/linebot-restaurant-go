package places

import (
	"encoding/json"
	"strconv"
)

// Place is main data struct
type Place struct {
	PlaceID      string  `json:"place_id" datastore:"place_id,noindex"`
	Name         string  `json:"name" datastore:"name,noindex"`
	Rating       float64 `json:"rating" datastore:"rating,noindex"`
	PhotoURI     string  `json:"photo_uri" datastore:"photo_uri,noindex"`
	GooglemapURI string  `json:"googlemap_uri" datastore:"googlemap_uri,noindex"`
}

// Places is Place slice
type Places []Place

// LatLng is used to unmarshal json
type LatLng struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

// UnmarshalJSON interface
func (ll *LatLng) UnmarshalJSON(b []byte) error {
	a := struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}{}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	ll.Lat = strconv.FormatFloat(a.Lat, 'f', -1, 64)
	ll.Lng = strconv.FormatFloat(a.Lng, 'f', -1, 64)
	return nil
}
