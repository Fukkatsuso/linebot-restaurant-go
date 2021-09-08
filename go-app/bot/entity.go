package bot

import (
	"cloud.google.com/go/datastore"
	mystore "github.com/Fukkatsuso/linebot-restaurant-go/go-app/datastore"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
)

// datastoreパッケージのEntityインターフェースを満たす構造体をここに定義する

// nearbySearchのクエリ
type Query struct {
	Lat      string   `json:"lat" datastore:"lat,noindex"`
	Lng      string   `json:"lng" datastore:"lng,noindex"`
	Keywords []string `json:"keywords" datastore:"keywords,noindex"`
	Radius   string   `json:"radius" datastore:"raduis,noindex"`
	Page     int      `json:"page" datastore:"page,noindex"`
}

func NewQuery(lat, lng string) Query {
	return Query{
		Lat:      lat,
		Lng:      lng,
		Keywords: []string{},
		Radius:   "500",
		Page:     0,
	}
}

func (query *Query) NameKey(name string, parent *datastore.Key) *datastore.Key {
	name = mystore.HashedString(name)
	return datastore.NameKey("Query", name, parent)
}

// ユーザのお気に入り
type Favorite struct {
	List []places.Place `datastore:"list,noindex"`
}

func (favorite *Favorite) NameKey(name string, parent *datastore.Key) *datastore.Key {
	name = mystore.HashedString(name)
	return datastore.NameKey("Favorite", name, parent)
}
