package datastore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"cloud.google.com/go/datastore"
)

// Entity is Datastore Resource
type Entity interface {
	NameKey(name string, parent *datastore.Key) *datastore.Key
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

// sha256でハッシュ化して64文字の文字列にする
func HashedString(base string) string {
	hashBytes := sha256.Sum256([]byte(base))
	hashString := hex.EncodeToString(hashBytes[:])
	return hashString
}
