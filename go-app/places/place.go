package places

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
