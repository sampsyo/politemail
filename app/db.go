package app

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

type Message struct {
	To      string
	Subject string
	Body    string
	Options []string
}

type User struct {
	Email      string
	MessageIds [][]byte
}

type Login struct {
	Email string
	Time  time.Time
}

func (a *App) addMessage(m *Message) []byte {
	key := []byte(uuid.New())
	j, _ := json.Marshal(m)
	a.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("messages"))
		err := b.Put(key, j)
		return err
	})
	return key
}

func (a *App) newLogin(email string) string {
	key := uuid.New()
	j, _ := json.Marshal(Login{
		email,
		time.Now(),
	})
	a.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("logins"))
		err := b.Put([]byte(key), j)
		return err
	})
	return key
}

func (a *App) getLogin(key string) (string, time.Time) {
	j := []byte{}
	a.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("logins"))
		j = b.Get([]byte(key))
		if j != nil {
			b.Delete([]byte(key))
		}
		return nil
	})
	if j == nil {
		return "", time.Now()
	}
	login := Login{}
	err := json.Unmarshal(j, &login)
	if err == nil {
		return login.Email, login.Time
	} else {
		log.Println("error unmarshalling login:", err)
		return "", time.Now()
	}
}
