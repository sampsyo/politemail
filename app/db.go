package app

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/boltdb/bolt"
)

func (a *App) addMessage(m *Message) []byte {
	key := []byte(uuid.New())
	a.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("messages"))
		j, _ := json.Marshal(m)
		err := b.Put(key, j)
		return err
	})
	return key
}
