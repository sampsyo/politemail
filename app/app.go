package app

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/mattbaird/gochimp"
	"github.com/stvp/go-toml-config"
	"log"
	"net/http"
	"path"
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
	router       *mux.Router
	baseURL      string
}

type ReqState struct {
	Email    string
	LoggedIn bool
}

const ReqStateKey int = 512

func getState(r *http.Request) ReqState {
	return context.Get(r, ReqStateKey).(ReqState)
}

func (a *App) getSession(r *http.Request) *sessions.Session {
	session, err := a.sessionStore.Get(r, "session")
	if err != nil {
		log.Fatal("invalid session:", err)
	}
	return session
}

func (a *App) redirectHome(w http.ResponseWriter, r *http.Request) {
	homeURL, _ := a.router.Get("home").URL()
	http.Redirect(w, r, homeURL.String(), 302)
}

type smap map[string]interface{}

func (a *App) render(w http.ResponseWriter, r *http.Request, template string,
	d smap) {
	if d == nil {
		d = smap{}
	}
	d["state"] = getState(r)
	d["logoutURL"], _ = a.router.Get("logout").URL()

	session := a.getSession(r)
	flashes := session.Flashes()
	d["flashes"] = flashes
	if len(flashes) > 0 {
		// Clear flashes.
		session.Save(r, w)
	}

	a.templates.Render(w, template, d)
}

func (a *App) beforeRequest(w http.ResponseWriter, r *http.Request) {
	session := a.getSession(r)
	email := session.Values["email"]
	context.Set(r, ReqStateKey, ReqState{
		Email:    email.(string),
		LoggedIn: email != "",
	})
}

func (a *App) Handler() http.Handler {
	if a.router == nil {
		r := mux.NewRouter()
		r.HandleFunc("/", a.handleHome).
			Name("home")
		r.HandleFunc("/compose", a.handleCompose)
		r.HandleFunc("/message", a.handleMessage)
		r.HandleFunc("/login/{key}", a.handleLoginCallback).
			Name("loginCallback")
		r.HandleFunc("/login", a.handleLogin)
		r.HandleFunc("/logout", a.handleLogout).
			Name("logout")

		// Probably just for debugging. Production should have this served by
		// a frontend.
		staticdir := path.Join(a.basedir, "static")
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticdir)))

		a.router = r
	}

	// Ridiculous-looking hack to get a "setup" call before each handler.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.beforeRequest(w, r)
		a.router.ServeHTTP(w, r)
	})
	return context.ClearHandler(h)
}

func New(basedir string, debug bool) *App {
	app := new(App)
	app.basedir = basedir

	// Read configuration.
	conf := config.NewConfigSet("", config.ExitOnError)
	mandrillKey := conf.String("mandrill_key", "")
	sessionKey := conf.String("secret_key", "")
	conf.StringVar(&app.adminFrom, "from", "politemail@example.com")
	conf.StringVar(&app.baseURL, "base_url", "")
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
