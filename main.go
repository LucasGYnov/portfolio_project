package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// "os"
	// "path/filepath"
	//"regexp"
	"text/template"

	data "portfolio_project/Data"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID       int
	Username string
	Password string
	Profile  string
	DB       *sql.DB
}

type Experience struct {
	ID      int
	Title   string
	Content string
}

type Contact struct {
	ID     int
	Numero int
	Email  string
	Postal string
}

type Formation struct {
	ID    int
	Title string
	Years int
}

type Tech struct {
	ID      int
	Title   string
	Content string
}

type MainPageData struct {
	IsLoggedIn     bool
	ProfilePicture string
	Experiences    []Experience
	Formations     []Formation
	Techs          []Tech
}

var (
	db       *sql.DB
	sessions = map[string]string{} // map to store session IDs and corresponding usernames
)

func main() {
	var err error
	db, err = data.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &mainPageHandler{})
	http.Handle("/login", &loginHandler{})
	http.Handle("/erreur", &errorHandler{})
	http.Handle("/logout", &logoutHandler{})
	http.Handle("/profil", &profilHandler{})
	http.Handle("/profilOther", &profilOtherHandler{})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))
	http.Handle("/img_video/", http.StripPrefix("/img_video/", http.FileServer(http.Dir("img_video/"))))

	fmt.Println("Serveur écoutant sur le port 6969...")
	log.Fatal(http.ListenAndServe("localhost:6969", nil))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type mainPageHandler struct{}

func (h *mainPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var data MainPageData
		sessionCookie, err := r.Cookie("session_id")
		if err == nil {
			username, ok := sessions[sessionCookie.Value]
			if ok {
				data.IsLoggedIn = true
				// Retrieve the profile picture of the user
				var profilePicture string
				err := db.QueryRow("SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture)
				if err == nil {
					data.ProfilePicture = profilePicture
				}
			}
		}

		// Retrieve posts with limit 7 and order by creation date
		rows, err := db.Query("SELECT e.id, e.title, e.content FROM experience e JOIN utilisateurs u ON e.id = u.id")
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des posts:", err)
			return
		}
		defer rows.Close()

		renderTemplate(w, "./src/index.html", data)
		return
	}
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type loginHandler struct{}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/login.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			setErrorCookie(w, "Username ou mot de passe vide")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		var dbPassword string
		err := db.QueryRow("SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				setErrorCookie(w, "Username ou mot de passe incorrect")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			setErrorCookie(w, "Erreur lors de la vérification de l'utilisateur")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			log.Println("Erreur lors de la vérification de l'utilisateur:", err)
			return
		}
		if password != dbPassword {
			setErrorCookie(w, "Mot de passe incorrect")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Créer une session
		sessionID := uuid.New().String()
		sessions[sessionID] = username
		cookie := &http.Cookie{
			Name:  "session_id",
			Value: sessionID,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type logoutHandler struct{}

func (h *logoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sessionCookie, err := r.Cookie("session_id")
		if err == nil {
			// Supprime la session du serveur
			delete(sessions, sessionCookie.Value)

			// Expire le cookie côté client
			sessionCookie.MaxAge = -1
			http.SetCookie(w, sessionCookie)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

func setErrorCookie(w http.ResponseWriter, errorMsg string) {
	cookie := &http.Cookie{
		Name:  "error",
		Value: errorMsg,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

type errorHandler struct{}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "./src/erreur.html", nil)
}

type profilHandler struct{}

func (h *profilHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username, ok := sessions[sessionCookie.Value]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user User
	err = db.QueryRow("SELECT id, username FROM utilisateurs WHERE username = ?", username).Scan(&user.ID, &user.Username)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des informations de l'utilisateur", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des informations de l'utilisateur:", err)
		return
	}

	renderTemplate(w, "./src/profil.html", user)
}

type profilOtherHandler struct{}

func (h *profilOtherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Nom d'utilisateur manquant", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow("SELECT id, username FROM utilisateurs WHERE username = ?", username).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
			return
		}
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération de l'utilisateur:", err)
		return
	}

	renderTemplate(w, "./src/profilOther.html", user)
}
