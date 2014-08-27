package app

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

type Message struct {
	From    string
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

// Add a user if it doesn't exist yet.
func (a *App) ensureUserExists(email string) {
	key := []byte(email)
	a.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		j := b.Get(key)
		if j == nil {
			j, _ := json.Marshal(User{
				email,
				[][]byte{},
			})
			return b.Put(key, j)
		}
		return nil
	})
}

func (a *App) addMessage(email string, m *Message) []byte {
	key := []byte(uuid.New())
	j, _ := json.Marshal(m)
	a.DB.Update(func(tx *bolt.Tx) error {
		// Create the message.
		messages := tx.Bucket([]byte("messages"))
		err := messages.Put(key, j)

		// Add the message to the user's list.
		users := tx.Bucket([]byte("users"))
		userj := users.Get([]byte(email))
		if userj == nil {
			return errors.New("no such user")
		}
		user := User{}
		err = json.Unmarshal(userj, &user)
		if err != nil {
			return err
		}
		user.MessageIds = append(user.MessageIds, key)
		newUserj, _ := json.Marshal(user)
		users.Put([]byte(email), newUserj)

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
