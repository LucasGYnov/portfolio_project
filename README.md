# Documentation Technique – Projet Portfolio

## Projet
**Développement d'une application full stack en 5 jours (Portfolio)**  
**Groupe :** Raphaël Caron, Jordan Milleville Lino, Oscar Li, Lucas Gerard, Yanis Fronteau

---

## Introduction

Ce projet a été réalisé dans un cadre de développement rapide, où l’objectif était de créer une application web Full Stack avec un délai restreint de 5 jours. Le but de l’application est de permettre à un utilisateur de gérer un portfolio personnel, incluant la modification d’informations telles que la photo de profil, le nom, les descriptions, et les projets via une interface d’administration.

Cette application a été développée en utilisant Go pour la gestion du back-end et une base de données SQLite3, tandis que le frontend est composé de pages HTML, CSS, et JavaScript. Le projet a permis d'expérimenter le travail collaboratif, la gestion de version avec GitHub, et l'application des bonnes pratiques de développement. L’application est conçue pour être facilement utilisable et personnalisable, tout en répondant aux attentes d’une interface moderne et responsive.

---

## Planification et Répartition des Rôles

La planification des tâches a été faite via des discussions de groupe, en utilisant des outils pour organiser les user stories et suivre les tâches dans un tableau Kanban. Chaque membre a choisi des responsabilités en fonction de ses compétences (backend, frontend, documentation, etc.). Cela a permis d'assurer une bonne répartition du travail.

### Répartition des tâches principales
- **Backend :** Jordan, Oscar, Raphaël
- **Frontend :** Lucas, Yanis
- **Documentation :** Raphaël, Lucas

---

## 1. Architecture de l'Application

L'architecture de l'application est basée sur une approche multi-page avec une gestion back-end en Go et une base de données SQLite3. Voici les éléments clés de l'architecture :

- **Front-end :** Composé de pages HTML, CSS et JavaScript avec un design inspiré d’un design d’actualité, le Bento Grid pour un portfolio moderne et épuré.

- **Pages principales :**
  - **Index :** Présente un portfolio avec la plupart des informations (Nom, description, Photos), des éléments cliquables permettant de voir les différents projets, formations, et réseaux sociaux.
  - **Formation :** Montre les différentes formations avec des informations utiles comme le nom de la formation, la période, ainsi qu’une description et photo facultative.
  - **Projets :** Affiche une liste de projets avec des informations détaillées (nom du projet, date, description, lien du projet, et photo facultative).
  - **Popup :** Page pour entrer les identifiants pour accéder à la page Admin.
  - **Admin :** Accessible uniquement aux utilisateurs autorisés, permettant de modifier le contenu du portfolio via des pop-ups et des formulaires.

- **Back-end :** Utilise Go pour gérer les requêtes (GET, POST, PUT, DELETE), interagissant avec la base de données SQLite3 pour mettre à jour les informations.

- **Base de données :** Le projet utilise SQLite3 pour stocker les informations utilisateur (nom, description, photos, etc.). Chaque mise à jour effectuée via l'interface admin est enregistrée dans la base de données et reflétée en HTML.

- **Sécurité :** Les mots de passe sont hashés pour sécuriser l'accès à la page d'administration.

---

## 2. Endpoints API (Go)

L'application utilise Go pour gérer les requêtes HTTP. Voici les principaux endpoints définis :

- **GET :** Récupère les informations à afficher dans les différentes pages du portfolio.
  - Exemple : `GET /api/user` pour récupérer les informations de l'utilisateur.
  
- **POST :** Permet l'envoi de nouvelles informations (ex. mise à jour de l'image ou des informations de description).
  - Exemple : `POST /api/update/image` pour envoyer une nouvelle image à enregistrer.
  
- **PUT :** Permet de mettre à jour les données existantes.
  - Exemple : `PUT /api/update/profile` pour modifier la description ou les informations personnelles de l'utilisateur.
  
- **DELETE :** Supprime des éléments si nécessaire, comme des images ou des projets.

Tous ces endpoints interagissent directement avec la base de données SQLite3.

---

## 3. Instructions pour Cloner, Installer et Exécuter l'Application

Voici les étapes pour cloner et exécuter l'application en local :

1. **Cloner le projet depuis GitHub :**  
   Ouvrir le terminal et exécuter la commande suivante :
   ```bash
   git clone https://github.com/LucasGYnov/portfolio_project.git
   ```

2. **Pré-requis :**
   - **Go :** Assurez-vous d'avoir Go installé sur votre machine. Si ce n'est pas le cas, vous pouvez l'installer à partir du site officiel : [Go](https://golang.org/dl/).
   - **SQLite3 :** SQLite3 est intégré, mais il est préférable de s'assurer qu'il est configuré correctement sur votre environnement.

3. **Installer les dépendances :**  
   Naviguer dans le dossier du projet :
   ```bash
   cd portfolio_project
   ```
   Exécuter le fichier Go :
   ```bash
   go run main.go
   ```

4. **Exécuter l'application :**  
   Une fois le serveur lancé, vous pouvez accéder à l'application via votre navigateur à l'adresse : `http://localhost:6969`.

---

## 4. Problème et Solution Implémentée

**Problème à Résoudre :**  
Créer une application de portfolio personnel permettant de modifier et d'actualiser facilement les informations du CV, y compris les images, descriptions et projets.

**Solution Implémentée :**  
L'interface admin permet à l'utilisateur de modifier les différentes sections de son portfolio via des formulaires, en gérant directement les informations dans la base de données SQLite3 via des requêtes Go. Cette interface est sécurisée par un système de connexion avec mots de passe hashés.

L’utilisateur accède à la page admin via l’URL `localhost:6969/admin`. Il doit se connecter avec un identifiant et mot de passe (ex. : `username : admin`, `mot de passe: admin` pour l’exemple). L’authentification est stockée pour chaque session. Si l’utilisateur n’a pas les identifiants corrects, il est renvoyé vers la page du formulaire pour réessayer.

---

## 5. Défis Rencontrés

Les principaux défis rencontrés incluent :
- **Gestion des merges :** Intégrer les différentes parties du code avec des délais serrés.
- **Travail en groupe :** La répartition des tâches et la coordination entre les membres ont nécessité une communication continue pour éviter les conflits et résoudre les problèmes techniques rencontrés, notamment lors des débogages collectifs.

---

## 6. Fonctionnalités de l'Application

L'application permet :
- De visualiser un portfolio via une interface moderne et responsive.
- De modifier les informations via une page admin sécurisée, incluant les images, textes, et réseaux sociaux.
- Hashing des mots de passe pour sécuriser l'accès à la page d'administration.
- Sauvegarde automatique des modifications dans une base de données SQLite3 et mise à jour instantanée du front-end.

---

## 7. Qualité du Code

- **Structure du code :** 
  - Go pour le back-end, avec une séparation claire des routes et des fichiers de gestion des requêtes.
  - Front-end : Organisé avec des fichiers HTML et CSS dans un dossier static, facilitant le chargement des images et des styles.
  
- **Commentaire :** Des commentaires ont été ajoutés pour les sections importantes, et des revues de code ont été réalisées entre les membres de l'équipe.

- **Pratiques standards :** Utilisation de commits réguliers et descriptifs, ainsi que de pull requests bien gérées.

---

## 8. Utilisation de GitHub

- Utilisation de plusieurs branches, dont une principale (`main`) et une dédiée au front-end (`front`).
- Les commits sont effectués régulièrement et incluent des descriptions claires des modifications.
- Les pull requests sont faites en équipe et discutées lors de sessions de travail ou via Discord pour assurer la synchronisation.

---

## 9. Travail d'Équipe

- **Répartition des tâches :** Selon les compétences et préférences de chaque membre.
- **Outils utilisés :** Notion pour suivre l'avancement (Kanban, tableau blanc) et Figma pour les maquettes.
- **Collaboration :** Partage de code sur Visual Studio Code, facilitant l'édition simultanée.
- **Communication :** Maintenue tout au long du projet avec des discussions régulières sur Discord.

---

## 10. Test et Validation

Des tests unitaires ont été effectués sur les routes back-end pour vérifier la gestion des requêtes (GET, POST, PUT, DELETE). 

- **Tests Front-end :** Tests manuels de chaque fonctionnalité sur différents navigateurs pour garantir une compatibilité cross-browser (Chrome, Firefox, Edge).
- **Tests Back-end :** Les endpoints ont été testés avec des outils comme Postman pour vérifier la bonne communication entre le front-end et le back-end.

---

## 11. Expériences Apprises

Ce projet a permis à l'équipe de renforcer certaines compétences techniques et collaboratives :
- **Gestion de version avec Git :** Amélioration dans la gestion

 des branches et la résolution des conflits.
- **Travail en équipe et communication :** Importance de la communication continue, notamment à travers Notion et Discord.
- **Debugging collectif :** Renforcement de la capacité à identifier rapidement les bugs et à y remédier efficacement.

---

## 12. Visuel de l’Application

VOIR PDF

---

## Conclusion

Ce projet a été une expérience enrichissante tant sur le plan technique que collaboratif. En seulement 5 jours, nous avons réussi à concevoir une application full stack fonctionnelle et sécurisée, avec une interface utilisateur moderne et une gestion simple des informations de portfolio via une page d'administration. Malgré les défis rencontrés, nous avons su trouver des solutions efficaces et garantir la qualité de l'application. Des améliorations futures sont possibles, notamment en termes de design et de fonctionnalités, mais nous sommes satisfaits du résultat atteint dans le temps imparti.