package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	// "os"
	// "path/filepath"
	"regexp"
	"text/template"

	data "portfolio_project/Data"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID       int
	Email    string
	Username string
	Password string
	Profile  string
	DB       *sql.DB
}

type Comment struct {
	ID       int
	PostID   int
	UserID   int
	Username string
	Content  string
}

type Post struct {
	ID       int
	Title    string
	Content  string
	Video    string
	Image    []string
	UserID   int
	Username string
	Comments []Comment
}

type MainPageData struct {
	IsLoggedIn     bool
	ProfilePicture string
	Posts          []Post
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
	http.Handle("/register", &registerHandler{})
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
			email, ok := sessions[sessionCookie.Value]
			if ok {
				data.IsLoggedIn = true
				// Retrieve the profile picture of the user
				var profilePicture string
				err := db.QueryRow("SELECT profile_picture FROM utilisateurs WHERE email = ?", email).Scan(&profilePicture)
				if err == nil {
					data.ProfilePicture = profilePicture
				}
			}
		}

		// Retrieve posts with limit 7 and order by creation date
		rows, err := db.Query("SELECT p.id, p.title, p.content, p.video, u.username FROM posts p JOIN utilisateurs u ON p.user_id = u.id ORDER BY p.created_at DESC LIMIT 7")
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des posts:", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var post Post
			var videoPtr sql.NullString
			err := rows.Scan(&post.ID, &post.Title, &post.Content, &videoPtr, &post.Username)
			if err != nil {
				http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
				log.Println("Erreur lors de la lecture des posts:", err)
				return
			}
			if videoPtr.Valid {
				post.Video = videoPtr.String
			}

			// Retrieve associated images
			imageRows, err := db.Query("SELECT image FROM posts WHERE post_id = ?", post.ID)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des images", http.StatusInternalServerError)
				log.Println("Erreur lors de la récupération des images:", err)
				return
			}
			defer imageRows.Close()

			for imageRows.Next() {
				var imagePath string
				if err := imageRows.Scan(&imagePath); err != nil {
					http.Error(w, "Erreur lors de la lecture des images", http.StatusInternalServerError)
					log.Println("Erreur lors de la lecture des images:", err)
					return
				}
				post.Image = append(post.Image, imagePath)
			}

			data.Posts = append(data.Posts, post)
		}

		renderTemplate(w, "./src/Main_page.html", data)
		return
	}
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type registerHandler struct{}

func (h *registerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/register.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setCookie(w, "error", "Erreur lors de la lecture du formulaire")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		if email == "" || username == "" || password == "" {
			setCookie(w, "error", "Email, nom d'utilisateur ou mot de passe vide")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		emailPattern := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
		matched, err := regexp.MatchString(emailPattern, email)
		if err != nil || !matched {
			setCookie(w, "error", "Email invalide")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		if usernameExists(username) {
			setCookie(w, "error", "Nom d'utilisateur deja pris, veuillez en choisir un autre")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		if emailExists(email) {
			setCookie(w, "error", "Email deja existente, veuillez en choisir un autre")
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		_, err = db.Exec("INSERT INTO utilisateurs (email, username, password) VALUES (?, ?, ?)", email, username, password)
		if err != nil {
			setCookie(w, "error", "Erreur lors de l'inscription")
			log.Println("Erreur lors de l'insertion dans la base de données:", err)
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

func setCookie(w http.ResponseWriter, name, value string) {
	cookie := &http.Cookie{
		Name:   name,
		Value:  value,
		Path:   "/",
		MaxAge: 10, // The cookie will be valid for 10 seconds
	}
	http.SetCookie(w, cookie)
}

func usernameExists(username string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM utilisateurs WHERE username = ?)"
	err := db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		log.Println("Erreur lors de la vérification du nom d'utilisateur :", err)
		return true
	}
	return exists
}

func emailExists(email string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM utilisateurs WHERE email = ?)"
	err := db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		log.Println("Erreur lors de la vérification du nom d'utilisateur :", err)
		return true
	}
	return exists
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
		email := r.FormValue("email")
		password := r.FormValue("password")
		if email == "" || password == "" {
			setErrorCookie(w, "Email ou mot de passe vide")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		var dbPassword string
		err := db.QueryRow("SELECT password FROM utilisateurs WHERE email = ?", email).Scan(&dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				setErrorCookie(w, "Email ou mot de passe incorrect")
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
		sessions[sessionID] = email
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
	email, ok := sessions[sessionCookie.Value]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user User
	err = db.QueryRow("SELECT id, email, username FROM utilisateurs WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Username)
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
    err := db.QueryRow("SELECT id, email, username FROM utilisateurs WHERE username = ?", username).Scan(&user.ID, &user.Email, &user.Username)
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