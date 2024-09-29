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
	Moi			[]Me
	Abouts      []About
	Contacts    []Contact
	Formations  []Formation
	Projects    []Project
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

        // Requête SQL pour récupérer les contacts
        rows, err := db.Query("SELECT instagram, twitter, behance, github, mail, linkedin FROM contact")
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

		// Récupérer le mot de passe haché de la base de données
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

		// Vérifier le mot de passe haché
		if !checkPasswordHash(password, dbPassword) {
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
		years := r.FormValue("years")
		link := r.FormValue("link")
		image := r.FormValue("image")
		instagram := r.FormValue("instagram")
		twitter := r.FormValue("twitter")
		behance := r.FormValue("behance")
		github := r.FormValue("github")
		mail := r.FormValue("mail")
		linkedin := r.FormValue("linkedin")

		// Insertion des données dans les tables correspondantes
		_, err = db.Exec("INSERT INTO me (title, content, user_id) VALUES (?, ?, ?)", title, content, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création de me", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table me :", err)
			return
		}

		// Insertion des données dans les tables correspondantes
		_, err = db.Exec("INSERT INTO about (content, user_id) VALUES (?, ?)", content, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création de l'expérience", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table me :", err)
			return
		}

		_, err = db.Exec("INSERT INTO contact (instagram, twitter, behance, github, mail, linkedin, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", instagram, twitter, behance, github, mail, linkedin, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création du contact", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table contact :", err)
			return
		}

		_, err = db.Exec("INSERT INTO formation (title, content, years, link, image, user_id) VALUES (?, ?, ?, ?, ?, ?)", title, content, years, link, image, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création de la formation", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table formation :", err)
			return
		}

		_, err = db.Exec("INSERT INTO project (title, content, years, link, image, user_id) VALUES (?, ?, ?, ?, ?, ?)", title, content, years, link, image, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création du project", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la table project :", err)
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

		// Récupération des me
		mes, err := fetchMes()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des me", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des me:", err)
			return
		}

		// Récupération des about
		abouts, err := fetchAbout()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des about", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des about:", err)
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

		// Récupération des project
		projects, err := fetchProjects()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des projects", http.StatusInternalServerError)
			log.Println("Erreur lors de la récupération des projects:", err)
			return
		}

		// Création d'un post avec les données récupérées
		post := Post{
			Moi: 		 mes,
			Abouts:      abouts,
			Contacts:    contacts,
			Formations:  formations,
			Projects:    projects,
		}

		// Ajout à la liste des posts
		posts = append(posts, post)

		// Rendu du template avec les posts
		renderTemplate(w, "./src/posts.html", posts)
		return
	}

	http.NotFound(w, r)
}

// Fonction pour récupérer les me depuis la base de données
func fetchMes() ([]Me, error) {
	rows, err := db.Query("SELECT id, title, content FROM me")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mes []Me
	for rows.Next() {
		var mois Me
		if err := rows.Scan(&mois.ID, &mois.Title, &mois.Content); err != nil {
			return nil, err
		}
		mes = append(mes, mois)
	}
	return mes, nil
}

// Fonction pour récupérer les about depuis la base de données
func fetchAbout() ([]About, error) {
	rows, err := db.Query("SELECT id, content FROM about")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ab []About
	for rows.Next() {
		var abs About
		if err := rows.Scan(&abs.ID, &abs.Content); err != nil {
			return nil, err
		}
		ab = append(ab, abs)
	}
	return ab, nil
}

// Fonction pour récupérer les contacts depuis la base de données
func fetchContacts() ([]Contact, error) {
	rows, err := db.Query("SELECT id, instagram, twitter, behance, github, mail, linkedin, image FROM contact")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		if err := rows.Scan(&contact.ID, &contact.Instagram, &contact.Twitter, &contact.Behance, &contact.Github, &contact.Mail, &contact.Linkedin); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// Fonction pour récupérer les formations depuis la base de données
func fetchFormations() ([]Formation, error) {
	rows, err := db.Query("SELECT id, title, content, years, link, image FROM formation")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var formations []Formation
	for rows.Next() {
		var formation Formation
		if err := rows.Scan(&formation.ID, &formation.Title, &formation.Content, &formation.Years, &formation.Link); err != nil {
			return nil, err
		}
		formations = append(formations, formation)
	}
	return formations, nil
}

// Fonction pour récupérer les project depuis la base de données
func fetchProjects() ([]Project, error) {
	rows, err := db.Query("SELECT id, title, content, years, link, image FROM project")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Title, &project.Content, &project.Years, &project.Link); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
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
	var me Me

	// Récupération des informations sur me
	err := db.QueryRow("SELECT m.id, m.title, m.content, u.username FROM me m JOIN utilisateurs u ON m.user_id = u.id WHERE m.id = ?", postID).
		Scan(&me.ID, &me.Title, &me.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post non trouvé", http.StatusNotFound)
			return
		}
		http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération du post:", err)
		return
	}

	// Récupération des informations de about
	var abs About
	err = db.QueryRow("SELECT a.id, a.content, u.username FROM about a JOIN utilisateurs u ON a.user_id = u.id WHERE a.id = ?", postID).
		Scan(&abs.ID,&abs.Content)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des about", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des about:", err)
		return
	}
	

	// Récupération des informations de contact
	var contact Contact
	err = db.QueryRow("SELECT c.id, c.instagram, c.twitter, c.behance, c.github, c.mail, c.linkedin, c.image, u.username FROM contact c JOIN utilisateurs u ON c.user_id = u.id WHERE c.id = ?", postID).
		Scan(&contact.ID, &contact.Instagram, &contact.Twitter, &contact.Behance, &contact.Github, &contact.Mail, &contact.Linkedin)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des contacts", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des contacts:", err)
		return
	}

	// Récupération des informations de formation
	var formation Formation
	err = db.QueryRow("SELECT f.id, f.title, f.content, f.years, f.link, f.image, u.username FROM formation f JOIN utilisateurs u ON f.user_id = u.id WHERE f.id = ?", postID).
		Scan(&formation.ID, &formation.Title, &formation.Content, &formation.Years, &formation.Link)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des formations", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des formations:", err)
		return
	}

	// Récupération des informations sur les project
	var project Project
	err = db.QueryRow("SELECT p.id, p.title, p.content, p.years, p.link, p.image, u.username FROM project p JOIN utilisateurs u ON p.user_id = u.id WHERE p.id = ?", postID).
		Scan(&project.ID, &project.Title, &project.Content, &project.Years, &project.Link)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur lors de la récupération des projects", http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération des projects:", err)
		return
	}

	// Ajout des autres informations au struct post
	post.Moi = []Me{}
	post.Abouts = []About{}
	post.Contacts = []Contact{}
	post.Formations = []Formation{}
	post.Projects = []Project{}

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
