package app

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mattbaird/gochimp"
	"log"
	"net/http"
	"time"
)

func (a *App) handleMessage(w http.ResponseWriter, r *http.Request) {
	state := getState(r)
	if !state.LoggedIn {
		http.Error(w, "not allowed", 403)
		return
	}

	email := getState(r).Email
	r.ParseForm()
	msg := Message{
		From:    email,
		To:      r.FormValue("to"),
		Subject: r.FormValue("subject"),
		Body:    r.FormValue("body"),
		Options: r.Form["option"],
	}
	a.addMessage(email, &msg)
	a.render(w, r, "confirm", smap{"message": msg})
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	state := getState(r)
	if state.LoggedIn {
		a.render(w, r, "compose", smap{"From": state.Email})
	} else {
		a.render(w, r, "login", nil)
	}
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	key := a.newLogin(email)

	callbackUrl, err := a.router.Get("loginCallback").URL("key", key)
	if err != nil {
		log.Fatal(err)
	}
	body := fmt.Sprintf(
		"Click this, please: %s%s",
		a.baseURL,
		callbackUrl,
	)

	m := gochimp.Message{
		Subject: "PoliteMail Login",
		Text:    body,
		To: []gochimp.Recipient{gochimp.Recipient{
			Email: email,
		}},
		FromEmail: a.adminFrom,
		FromName:  "PoliteMail",
	}

	log.Println("sending login email to", email)
	_, err = a.mandrill.MessageSend(m, false)
	session := a.getSession(r)
	if err == nil {
		log.Println("login email sent successfully")
		session.AddFlash("Login email sent. Please check your mail.")
	} else {
		log.Println("loging email failed:", err)
		session.AddFlash("The login email failed to send!")
	}
	session.Save(r, w)
	a.redirectHome(w, r)
}

func (a *App) verifyLogin(key string) (string, error) {
	email, loginTime := a.getLogin(key)
	if email == "" {
		return "", errors.New("login request not found")
	}
	ago := time.Since(loginTime)
	if ago > time.Hour {
		return "", errors.New(
			fmt.Sprintf("login too old: %i", ago),
		)
	}
	return email, nil
}

func (a *App) handleLoginCallback(w http.ResponseWriter, r *http.Request) {
	session := a.getSession(r)

	key := mux.Vars(r)["key"]
	log.Println("verifying login", key)
	email, err := a.verifyLogin(key)
	if err == nil {
		log.Println("login verified")
		session.Values["email"] = email
		a.ensureUserExists(email)
		session.AddFlash("You are now logged in.")
	} else {
		log.Println("login failed:", err)
		session.AddFlash("This login link is invalid. Please try again.")
	}

	session.Save(r, w)
	a.redirectHome(w, r)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	session := a.getSession(r)
	session.Values["email"] = ""
	session.AddFlash("You are now logged out.")
	session.Save(r, w)
	a.redirectHome(w, r)
}
