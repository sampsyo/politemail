package app

import (
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
)

type Message struct {
	To      string
	Subject string
	Body    string
	Options []string
}

type User struct {
	Email      string
	Password   []byte
	MessageIds [][]byte
}

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

func (u *User) SetPassword(password string) {
	p, err := bcrypt.GenerateFromPassword([]byte(password),
		bcrypt.DefaultCost)
	if err != nil {
		log.Panic(err)
	}
	u.Password = p
}

func (u *User) TestPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	return err == nil
}
