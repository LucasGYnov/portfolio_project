package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// "os"
	// "path/filepath"
	// "regexp"
	"text/template"

	data "portfolio_project/Data"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Password string
	Profile  string
	DB       *sql.DB
}

type Post struct {
	ID          int
	Moi			[]Me
	Abouts      []About
	Contacts    []Contact
	Formations  []Formation
	Projects    []Project
}

type Me struct {
	ID      int
	Title   string
	Content string
}

type About struct {
	ID      int
	Content string
}

type Contact struct {
	ID     int
	Instagram string
	Twitter  string
	Behance string
	Github string
	Mail string
	Linkedin string
}

type Formation struct {
	ID    int
	Title string
	Content string
	Years int
	Link  string
	Image []string
}

type Project struct {
	ID    int
	Title string
	Content string
	Years int
	Link  string
	Image []string
}

type MainPageData struct {
	IsLoggedIn     bool
	ProfilePicture string
	Posts          []Post
	Users		   []User
	Moi			   []Me
	Abouts         []About
	Contacts       []Contact
	Formations     []Formation
	Projects       []Project
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

	createAdmin()

	http.Handle("/", &mainPageHandler{})
	http.Handle("/erreur", &errorHandler{})
	http.Handle("/popup", &popupHandler{})
	http.Handle("/admin", &adminHandler{})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))

	fmt.Println("Serveur écoutant sur le port 6969...")
	log.Fatal(http.ListenAndServe("localhost:6969", nil))
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func createAdmin() {
	var username string
	err := db.QueryRow("SELECT username FROM utilisateurs WHERE username = 'admin'").Scan(&username)
	if err == sql.ErrNoRows {
		// Si l'utilisateur admin n'existe pas, on le crée
		hashedPassword, err := hashPassword("admin")
		if err != nil {
			log.Fatal("Erreur lors du hac hage du mot de passe:", err)
		}

		_, err = db.Exec("INSERT INTO utilisateurs (username, password) VALUES (?, ?)", "admin", hashedPassword)
		if err != nil {
			log.Fatal("Erreur lors de la création de l'utilisateur admin:", err)
		}

		fmt.Println("Utilisateur admin créé avec succès")
	} else if err != nil {
		log.Fatal("Erreur lors de la vérification de l'utilisateur admin:", err)
	} else {
		fmt.Println("L'utilisateur admin existe déjà")
	}
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

		// Requête SQL pour récupérer les about
        rows, err := db.Query("SELECT username, password FROM utilisateurs")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des users", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des users:", err)
            return
        }
        defer rows.Close()

        var users []User
        for rows.Next() {
            var user User
            if err := rows.Scan(&user.Username, &user.Password); err != nil {
                http.Error(w, "Erreur lors du scan des users", http.StatusInternalServerError)
                log.Println("Erreur lors du scan des users:", err)
                return
            }
            users = append(users, user)
        }

        data.Users = users // Assignez les users à MainPageData

		// Requête SQL pour récupérer les about
        rows, err = db.Query("SELECT content FROM about")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des about", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des about:", err)
            return
        }
        defer rows.Close()

        var abouts []About
        for rows.Next() {
            var about About
            if err := rows.Scan(&about.Content); err != nil {
                http.Error(w, "Erreur lors du scan des about", http.StatusInternalServerError)
                log.Println("Erreur lors du scan des about:", err)
                return
            }
            abouts = append(abouts, about)
        }

        data.Abouts = abouts // Assignez les about à MainPageData

        // Requête SQL pour récupérer les contacts
        rows, err = db.Query("SELECT instagram, twitter, behance, github, mail, linkedin FROM contact")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des contacts", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des contacts:", err)
            return
        }
        defer rows.Close()

        var contacts []Contact
        for rows.Next() {
            var contact Contact
            if err := rows.Scan(&contact.Instagram, &contact.Twitter, &contact.Behance, &contact.Github, &contact.Mail, &contact.Linkedin); err != nil {
                http.Error(w, "Erreur lors du scan des contacts", http.StatusInternalServerError)
                log.Println("Erreur lors du scan des contacts:", err)
                return
            }
            contacts = append(contacts, contact)
        }

        data.Contacts = contacts // Assignez les contacts à MainPageData

		// Requête SQL pour récupérer les me
        rows, err = db.Query("SELECT title, content, years, link FROM formation")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des formation", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des formation:", err)
            return
        }
        defer rows.Close()

        var formations []Formation
        for rows.Next() {
            var formation Formation
            if err := rows.Scan(&formation.Title, &formation.Content, &formation.Years, &formation.Link); err != nil {
                http.Error(w, "Erreur lors du scan des formation", http.StatusInternalServerError)
                log.Println("Erreur lors du scan des formation:", err)
                return
            }
            formations = append(formations, formation)
        }

        data.Formations = formations // Assignez les contacts à MainPageData

		// Requête SQL pour récupérer les me
        rows, err = db.Query("SELECT title, content, years, link FROM project")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des project", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des project:", err)
            return
        }
        defer rows.Close()

        var projects []Project
        for rows.Next() {
            var project Project
            if err := rows.Scan(&project.Title, &project.Content, &project.Years, &project.Link); err != nil {
                http.Error(w, "Erreur lors du scan des project", http.StatusInternalServerError)
                log.Println("Erreur lors du scan des project:", err)
                return
            }
            projects = append(projects, project)
        }

        data.Projects = projects // Assignez les contacts à MainPageData

		// Requête SQL pour récupérer les me
        rows, err = db.Query("SELECT title, content FROM me")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des me", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des me:", err)
            return
        }
        defer rows.Close()

        var mois []Me
        for rows.Next() {
            var moi Me
            if err := rows.Scan(&moi.Title, &moi.Content); err != nil {
                http.Error(w, "Erreur lors du scan des me", http.StatusInternalServerError)
                log.Println("Erreur lors du scan des me:", err)
                return
            }
            mois = append(mois, moi)
        }

        data.Moi = mois // Assignez les me à MainPageData


        renderTemplate(w, "./src/index.html", data)
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

type popupHandler struct{}

func (h *popupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/popup.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		if password == "" {
			setErrorCookie(w, "Username ou mot de passe vide")
			http.Redirect(w, r, "/popup", http.StatusSeeOther)
			return
		}

		var dbPassword string
		err := db.QueryRow("SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				setErrorCookie(w, "Username ou mot de passe incorrect")
				http.Redirect(w, r, "/popup", http.StatusSeeOther)
				return
			}
			setErrorCookie(w, "Erreur lors de la vérification de l'utilisateur")
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			log.Println("Erreur lors de la vérification de l'utilisateur:", err)
			return
		}

		// Use the checkPasswordHash function here
		if !checkPasswordHash(password, dbPassword) {
			setErrorCookie(w, "Mot de passe incorrect")
			http.Redirect(w, r, "/popup", http.StatusSeeOther)
			return
		}

		// Create a session
		sessionID := uuid.New().String()
		sessions[sessionID] = username
		cookie := &http.Cookie{
			Name:  "session_id",
			Value: sessionID,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type adminHandler struct{}

func (h *adminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Vérifie si une session est active
	if cookie, err := r.Cookie("session_id"); err != nil || sessions[cookie.Value] == "" {
		// Si aucune session valide n'existe, redirige vers la page de connexion
		http.Redirect(w, r, "/popup", http.StatusSeeOther)
		return
	}

	// Si la session est valide, affiche la page admin
	renderTemplate(w, "./src/admin.html", nil)
}