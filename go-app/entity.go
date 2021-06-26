package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/Fukkatsuso/linebot-restaurant-go/go-app/places"
)

// Entity is Datastore Resource
type Entity interface {
	NameKey(name string, parent *datastore.Key) *datastore.Key
}

// sha256でハッシュ化して64文字の文字列にする
func HashedString(base string) string {
	hashBytes := sha256.Sum256([]byte(base))
	hashString := hex.EncodeToString(hashBytes[:])
	return hashString
}

// Query inplements Entity interface
type Query struct {
	Lat      string   `json:"lat" datastore:"lat,noindex"`
	Lng      string   `json:"lng" datastore:"lng,noindex"`
	Keywords []string `json:"keywords" datastore:"keywords,noindex"`
	Radius   string   `json:"radius" datastore:"raduis,noindex"`
	Page     int      `json:"page" datastore:"page,noindex"`
}

// NewQuery is NearbySearch's Query
func NewQuery(lat, lng string) Query {
	return Query{
		Lat:      lat,
		Lng:      lng,
		Keywords: []string{},
		Radius:   "500",
		Page:     0,
	}
}

// NameKey returns Query's Datastore-Key
func (query *Query) NameKey(name string, parent *datastore.Key) *datastore.Key {
	name = HashedString(name)
	return datastore.NameKey("Query", name, parent)
}

// Favorite inplements Entity interface
type Favorite struct {
	List []places.Place `datastore:"list,noindex"`
}

// NameKey returns Favorite's Datastore-Key
func (favorite *Favorite) NameKey(name string, parent *datastore.Key) *datastore.Key {
	name = HashedString(name)
	return datastore.NameKey("Favorite", name, parent)
}

// Get Entity from Datastore
func Get(ctx context.Context, client *datastore.Client, entity Entity, name string, parent *datastore.Key) error {
	key := entity.NameKey(name, parent)
	err := client.Get(ctx, key, entity)
	log.Println("[Get]", entity, err)
	return err
}

// Save Entity in Datastore
func Save(ctx context.Context, client *datastore.Client, entity Entity, name string, parent *datastore.Key) error {
	key := entity.NameKey(name, parent)
	_, err := client.Put(ctx, key, entity)
	log.Println("[Save]", entity, err)
	return err
}

// Delete Entity from Datastore
func Delete(ctx context.Context, client *datastore.Client, entity Entity, name string, parent *datastore.Key) error {
	key := entity.NameKey(name, parent)
	err := client.Delete(ctx, key)
	log.Println("[Delete]", entity, err)
	return err
}
