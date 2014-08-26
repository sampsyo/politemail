package app

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/mattbaird/gochimp"
	"github.com/stvp/go-toml-config"
	"log"
	"net/http"
	"path"
	"time"
)

import (
	"github.com/sampsyo/politemail/tmplpool"
)

type App struct {
	basedir      string
	templates    *tmplpool.Pool
	DB           *bolt.DB
	mandrill     *gochimp.MandrillAPI
	adminFrom    string
	sessionStore *sessions.CookieStore
}

func (a *App) getSession(r *http.Request) *sessions.Session {
	session, err := a.sessionStore.Get(r, "session")
	if err != nil {
		log.Fatal("invalid session:", err)
	}
	return session
}

func (a *App) handleCompose(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Foo string
	}{
		"bar",
	}
	a.templates.Render(w, "compose", data)
}

func (a *App) handleMessage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	msg := Message{
		r.FormValue("to"),
		r.FormValue("subject"),
		r.FormValue("body"),
		r.Form["option"],
	}
	a.addMessage(&msg)
	a.templates.Render(w, "confirm", msg)
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	session := a.getSession(r)
	email, found := session.Values["email"]
	if found {
		email_ := email.(string)
		a.templates.Render(w, "compose", struct{ From string }{email_})
	} else {
		a.templates.Render(w, "login", nil)
	}
}

type Status struct {
	Title string
	Body  string
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	key := a.newLogin(email)

	body := fmt.Sprintf(
		"Click this, please: http://0.0.0.0:8080/login/%s",
		key,
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
	_, err := a.mandrill.MessageSend(m, false)
	if err == nil {
		log.Println("login email sent successfully")
		a.templates.Render(w, "status", Status{
			"Login Email Sent",
			"Take a look!",
		})
	} else {
		log.Println("loging email failed:", err)
		a.templates.Render(w, "status", Status{
			"Login Failed",
			"Sorry. :(",
		})
	}
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
	key := mux.Vars(r)["key"]
	log.Println("verifying login", key)
	email, err := a.verifyLogin(key)
	if err == nil {
		log.Println("login verified")

		session := a.getSession(r)
		session.Values["email"] = email
		session.Save(r, w)

		a.templates.Render(w, "status", Status{
			"Login Successful",
			fmt.Sprintf(
				"You are logged in as %s.",
				email,
			),
		})
	} else {
		log.Println("login failed:", err)
		a.templates.Render(w, "status", Status{
			"Login Failed",
			"Try again, yo.",
		})
	}
}

func (a *App) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", a.handleHome)
	r.HandleFunc("/compose", a.handleCompose)
	r.HandleFunc("/message", a.handleMessage)
	r.HandleFunc("/login/{key}", a.handleLoginCallback)
	r.HandleFunc("/login", a.handleLogin)
	staticdir := path.Join(a.basedir, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticdir)))
	return context.ClearHandler(r)
}

func New(basedir string, debug bool) *App {
	app := new(App)
	app.basedir = basedir

	// Read configuration.
	conf := config.NewConfigSet("", config.ExitOnError)
	mandrillKey := conf.String("mandrill_key", "")
	sessionKey := conf.String("secret_key", "")
	conf.StringVar(&app.adminFrom, "from", "politemail@example.com")
	err := conf.Parse(path.Join(basedir, "config.toml"))
	if err != nil {
		log.Fatal(err)
	}

	// Template pool.
	app.templates = tmplpool.New(path.Join(basedir, "template"))
	app.templates.Debug = debug
	app.templates.Common = []string{"base"}
	app.templates.BaseDef = "base"

	// Database connection.
	db, err := bolt.Open(path.Join(basedir, "politemail.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("messages"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("logins"))
		return err
	})
	app.DB = db

	// Mandrill API.
	if *mandrillKey == "" {
		log.Fatal("no Mandrill key in config")
	}
	mandrill, err := gochimp.NewMandrill(*mandrillKey)
	if err != nil {
		log.Fatal(err)
	}
	app.mandrill = mandrill

	// Session store.
	if *sessionKey == "" {
		log.Fatal("no secret key in config")
	}
	app.sessionStore = sessions.NewCookieStore([]byte(*sessionKey))

	return app
}

func (a *App) Teardown() {
	a.DB.Close()
}
