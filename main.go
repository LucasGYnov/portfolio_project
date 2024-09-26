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
)

type User struct {
	ID       int
	Username string
	Password string
	Profile  string
	DB       *sql.DB
}

type Post struct {
	ID       int
	Experiences    []Experience
	Contacts 	   []Contact
	Formations     []Formation
	Techs          []Tech
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
	http.Handle("/login", &loginHandler{})
	http.Handle("/newpost", &newPostHandler{})
	http.Handle("/posts", &postsHandler{})
	http.Handle("/details/", &postDetailHandler{})
	http.Handle("/erreur", &errorHandler{})
	http.Handle("/logout", &logoutHandler{})
	http.Handle("/popup", &popupHandler{})
	http.Handle("/admin", &adminHandler{})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))

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

		// Assurez-vous que cette requête SQL est conforme à la structure de votre base de données
		rows, err := db.Query("SELECT e.id, c.id, f.id, t.id, u.username FROM experience e JOIN utilisateurs u ON e.exp_id = u.id JOIN contact c ON c.contact_id = u.id JOIN formation f ON f.formation_id = u.id JOIN tech t ON t.tech_id = u.id")
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des posts:", err)
			return
		}
		defer rows.Close()
		renderTemplate(w, "./src/index.html", data)
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

type newPostHandler struct{}

func (h *newPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/new_post.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		// Vérifier si l'utilisateur est connecté
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
		var userID int
		err = db.QueryRow("SELECT id FROM utilisateurs WHERE username = ?", username).Scan(&userID)
		if err != nil {
			http.Error(w, "Erreur lors de la vérification de l'utilisateur", http.StatusInternalServerError)
			log.Println("Erreur lors de la vérification de l'utilisateur:", err)
			return
		}

		// Gestion de l'envoi du formulaire
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		content := r.FormValue("content")
		numero := r.FormValue("numero")
		email := r.FormValue("email")
		postal := r.FormValue("postal")
		years := r.FormValue("years")

		// Insertion des données dans les tables correspondantes
		_, err = db.Exec("INSERT INTO experience (title, content, exp_id) VALUES (?, ?, ?)", title, content, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création de l'expérience", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table experience :", err)
			return
		}

		_, err = db.Exec("INSERT INTO contact (numero, email, postal, contact_id) VALUES (?, ?, ?, ?)", numero, email, postal, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création du contact", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table contact :", err)
			return
		}

		_, err = db.Exec("INSERT INTO formation (title, content, years, formation_id) VALUES (?, ?, ?, ?)", title, content, years, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création de la formation", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table formation :", err)
			return
		}

		_, err = db.Exec("INSERT INTO tech (title, content, tech_id) VALUES (?, ?, ?)", title, content, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création de la technologie", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table tech :", err)
			return
		}

		// Redirection vers la page principale
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.NotFound(w, r)
}


type postsHandler struct{}

func (h *postsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var posts []Post

		// Récupération des expériences
		experiences, err := fetchExperiences()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des expériences", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des expériences:", err)
			return
		}

		// Récupération des contacts
		contacts, err := fetchContacts()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des contacts", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des contacts:", err)
			return
		}

		// Récupération des formations
		formations, err := fetchFormations()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des formations", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des formations:", err)
			return
		}

		// Récupération des technologies
		techs, err := fetchTechs()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des technologies", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des technologies:", err)
			return
		}

		// Création d'un post avec les données récupérées
		post := Post{
			Experiences: experiences,
			Contacts:    contacts,
			Formations:  formations,
			Techs:       techs,
		}

		// Ajout à la liste des posts
		posts = append(posts, post)

		// Rendu du template avec les posts
		renderTemplate(w, "./src/posts.html", posts)
		return
	}

	http.NotFound(w, r)
}

// Fonction pour récupérer les expériences depuis la base de données
func fetchExperiences() ([]Experience, error) {
	rows, err := db.Query("SELECT id, title, content FROM experience")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []Experience
	for rows.Next() {
		var exp Experience
		if err := rows.Scan(&exp.ID, &exp.Title, &exp.Content); err != nil {
			return nil, err
		}
		experiences = append(experiences, exp)
	}
	return experiences, nil
}

// Fonction pour récupérer les contacts depuis la base de données
func fetchContacts() ([]Contact, error) {
	rows, err := db.Query("SELECT id, numero, email, postal FROM contact")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		if err := rows.Scan(&contact.ID, &contact.Numero, &contact.Email, &contact.Postal); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// Fonction pour récupérer les formations depuis la base de données
func fetchFormations() ([]Formation, error) {
	rows, err := db.Query("SELECT id, title, years FROM formation")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formations []Formation
	for rows.Next() {
		var formation Formation
		if err := rows.Scan(&formation.ID, &formation.Title, &formation.Years); err != nil {
			return nil, err
		}
		formations = append(formations, formation)
	}
	return formations, nil
}

// Fonction pour récupérer les technologies depuis la base de données
func fetchTechs() ([]Tech, error) {
	rows, err := db.Query("SELECT id, title, content FROM tech")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techs []Tech
	for rows.Next() {
		var tech Tech
		if err := rows.Scan(&tech.ID, &tech.Title, &tech.Content); err != nil {
			return nil, err
		}
		techs = append(techs, tech)
	}
	return techs, nil
}

type postDetailHandler struct{}

func (h *postDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Récupération du postID depuis l'URL
	postID := r.URL.Path[len("/details/"):]
	if postID == "" {
		http.Error(w, "ID du post manquant dans l'URL", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		// Gestion de la soumission de nouveaux commentaires
		sessionCookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Vérification de la session
		username, ok := sessions[sessionCookie.Value]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Récupération de l'ID utilisateur
		var userID int
		err = db.QueryRow("SELECT id FROM utilisateurs WHERE username = ?", username).Scan(&userID)
		if err != nil {
			http.Error(w, "Erreur lors de la vérification de l'utilisateur", http.StatusInternalServerError)
			log.Println("Erreur lors de la vérification de l'utilisateur:", err)
			return
		}

		// Redirection après l'ajout réussi d'un commentaire
		http.Redirect(w, r, fmt.Sprintf("/details/%s", postID), http.StatusSeeOther)
		return
	}

	var post Post

	// Récupération des détails du post et des commentaires pour la requête GET
	var experience Experience

	// Récupération des informations sur l'expérience
	err := db.QueryRow("SELECT e.id, e.title, e.content, u.username FROM experience e JOIN utilisateurs u ON e.exp_id = u.id WHERE e.id = ?", postID).
		Scan(&experience.ID, &experience.Title, &experience.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post non trouvé", http.StatusNotFound)
			return
		}
		http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération du post:", err)
		return
	}

	// Récupération des informations de contact
	var contact Contact
	err = db.QueryRow("SELECT c.id, c.numero, c.email, c.postal, u.username FROM contact c JOIN utilisateurs u ON c.contact_id = u.id WHERE c.id = ?", postID).
		Scan(&contact.ID, &contact.Numero, &contact.Email, &contact.Postal)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des contacts", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des contacts:", err)
		return
	}

	// Récupération des informations de formation
	var formation Formation
	err = db.QueryRow("SELECT f.id, f.title, f.years, u.username FROM formation f JOIN utilisateurs u ON f.formation_id = u.id WHERE f.id = ?", postID).
		Scan(&formation.ID, &formation.Title, &formation.Years)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des formations", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des formations:", err)
		return
	}

	// Récupération des informations sur les technologies
	var tech Tech
	err = db.QueryRow("SELECT t.id, t.title, t.content, u.username FROM tech t JOIN utilisateurs u ON t.tech_id = u.id WHERE t.id = ?", postID).
		Scan(&tech.ID, &tech.Title, &tech.Content)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des technologies", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des technologies:", err)
		return
	}

	// Ajout des autres informations au struct post
	post.Contacts = []Contact{}
	post.Formations = []Formation{}
	post.Techs = []Tech{}

	// Rendu du template avec les données récupérées
	renderTemplate(w, "./src/post_detail.html", post)
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
        if password != dbPassword {
            setErrorCookie(w, "Mot de passe incorrect")
            http.Redirect(w, r, "/popup", http.StatusSeeOther)
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
