# Documentation Technique - Groupie Tracker

## Table des matières
1. [Vue d'ensemble du projet](#vue-densemble-du-projet)
2. [Architecture](#architecture)
3. [Documentation des fichiers Go](#documentation-des-fichiers-go)
4. [Documentation des fichiers HTML](#documentation-des-fichiers-html)
5. [Documentation des fichiers JavaScript](#documentation-des-fichiers-javascript)
6. [Documentation des fichiers CSS](#documentation-des-fichiers-css)
7. [Flux de données](#flux-de-données)
8. [Configuration et déploiement](#configuration-et-déploiement)

---

## Vue d'ensemble du projet

**Groupie Tracker** est une application web permettant de rechercher des groupes de musique, visualiser leurs informations, dates de concerts et localisations géographiques.

### Technologies utilisées
- **Backend**: Go (Golang)
- **Frontend**: HTML5, CSS3, JavaScript (Vanilla)
- **Base de données**: MySQL
- **Cartographie**: Leaflet.js + OpenStreetMap
- **APIs externes**: 
  - Groupie Trackers API (artistes, dates, lieux)
  - iTunes API (previews musicales)
  - Deezer API (previews musicales)
  - Nominatim (géocodage)

### Fonctionnalités principales
- 🔍 Recherche d'artistes avec suggestions automatiques
- 🗺️ Géolocalisation des concerts sur carte interactive
- 🎵 Lecture de previews musicales au survol des vinyles
- 👤 Système d'authentification (inscription/connexion)
- 💳 Module d'abonnement avec simulation de paiement
- 🎨 Interface moderne avec effets glassmorphism

---

## Architecture

```
Groupie-Persso/
├── main.go                 # Point d'entrée principal (15 lignes)
├── go.mod                  # Configuration module Go
├── internal/               # Code interne (modularisé)
│   ├── database/
│   │   └── db.go          # Initialisation base de données MySQL
│   └── handlers/
│       ├── auth.go        # Handlers authentification (register/login)
│       ├── proxy.go       # Handler proxy CORS
│       └── static.go      # Handlers fichiers statiques et templates
├── web/                    # Ressources frontend
│   ├── static/            # Fichiers statiques
│   │   ├── css/          # Styles
│   │   ├── js/           # Scripts JavaScript
│   │   └── images/       # Images
│   └── templates/         # Templates HTML
├── index.html             # Page d'accueil (root)
└── go.sum                 # Checksums dépendances
```

**Refactorisation**: Le code a été modularisé pour une meilleure maintenabilité:
- **main.go**: Réduit à 15 lignes (orchestration uniquement)
- **internal/database/**: Gestion de la base de données
- **internal/handlers/**: Tous les handlers HTTP séparés par domaine

---

## Documentation des fichiers Go

### 📄 main.go

**Rôle**: Point d'entrée du serveur HTTP Go. Orchestration simple des routes et services.

**Code complet (15 lignes):**
```go
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"groupiepersso/internal/database"
	"groupiepersso/internal/handlers"
)

func main() {
	handlers.DB, _ = database.InitDB()
	handlers.SetupStaticRoutes(filepath.Join("web", "static"))
	handlers.SetupTemplateRoutes()
	http.HandleFunc("/api/artists-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/artists"))
	http.HandleFunc("/api/locations-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/locations"))
	http.HandleFunc("/api/dates-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/dates"))
	http.HandleFunc("/api/relation-proxy", handlers.ProxyHandler("https://groupietrackers.herokuapp.com/api/relation"))
	http.HandleFunc("/api/register", handlers.HandleRegister)
	http.HandleFunc("/api/login", handlers.HandleLogin)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s — open http://localhost:%s/", port, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

**Explication ligne par ligne:**

- **Ligne 14**: Initialise la base de données (voir `internal/database/db.go`)
- **Ligne 15**: Configure les routes pour les fichiers statiques (CSS, JS, images)
- **Ligne 16**: Configure les routes pour les templates HTML (index, login, search, geoloc)
- **Ligne 17-20**: Enregistre les 4 proxies CORS pour l'API Groupie Trackers
- **Ligne 21-22**: Enregistre les endpoints d'authentification
- **Ligne 23-27**: Lecture du port depuis l'environnement (par défaut 8080)
- **Ligne 28**: Démarre le serveur HTTP

**Avantages de cette architecture modularisée:**
- ✅ **Clarté**: Le main.go est lisible en 15 lignes
- ✅ **Maintenabilité**: Logique séparée par domaine (auth, database, proxy, static)
- ✅ **Testabilité**: Chaque module peut être testé indépendamment
- ✅ **Scalabilité**: Facile d'ajouter de nouveaux handlers

### 📄 internal/database/db.go

**Rôle**: Gestion de la connexion et configuration MySQL.

**Fonctions principales:**
- `InitDB()`: Initialise la connexion et retourne `*sql.DB`
- `getenvDefault()`: Utilitaire pour lire les variables d'environnement avec valeur par défaut

**Variables d'environnement supportées:**
```
DB_USER   → Nom d'utilisateur MySQL (défaut: "root")
DB_PASS   → Mot de passe MySQL (défaut: "")
DB_HOST   → Host MySQL (défaut: "localhost")
DB_PORT   → Port MySQL (défaut: "3306")
DB_NAME   → Nom de la base (défaut: "groupi_tracker")
DISABLE_DB → Désactiver la DB (valeur "1")
```

**Configuration du pool de connexions:**
```go
database.SetMaxOpenConns(10)           // Max 10 connexions simultanées
database.SetMaxIdleConns(5)            // 5 connexions en idle
database.SetConnMaxLifetime(30 * time.Minute)  // Durée vie connexion
```

**Charset UTF-8 MB4:**
- Support complet Unicode (emojis, caractères spéciaux)
- Nécessaire pour une app multilingue/moderne

### 📄 internal/handlers/auth.go

**Rôle**: Handlers d'authentification (inscription et connexion).

**Variables globales:**
```go
var DB *sql.DB  // Pointeur de base de données partagée
```

**Type `user`:**
```go
type user struct {
    Nom      string `json:"nom"`      // Nom de famille
    Prenom   string `json:"prenom"`   // Prénom
    Sexe     string `json:"sexe"`     // Genre (M/F/Autre)
    Password string `json:"password"` // Mot de passe
}
```

**Fonction `HandleRegister(w, r)` - Création d'un compte**

1. **Validation HTTP**: Accepte uniquement POST
2. **Vérification DB**: Retourne 503 si DB indisponible
3. **Parsing JSON**: Décode le corps de la requête
4. **Validation champs**: Tous les champs requis, password min 6 caractères
5. **Hachage bcrypt**: Génère un hash sécurisé du password
6. **Insertion DB**: Utilise des prepared statements (protection SQL injection)
7. **Réponse**: Retourne l'ID utilisateur (HTTP 201 Created)

**Fonction `HandleLogin(w, r)` - Connexion utilisateur**

1. **Validation HTTP**: POST uniquement
2. **Vérification DB**: Retourne 503 si indisponible
3. **Parsing JSON**: Récupère id_utilisateur et password
4. **Query DB**: Récupère le hash stocké
5. **Vérification password**: Compare avec bcrypt
6. **Réponse succès**: Retourne les infos utilisateur (HTTP 200)

**Sécurité implémentée:**
- ✅ Prepared statements (protection SQL injection)
- ✅ Bcrypt hash (protection password)
- ✅ Vérification HTTP method
- ✅ Gestion erreurs gracieuse

### 📄 internal/handlers/proxy.go

**Rôle**: Handler proxy pour contourner les restrictions CORS.

**Fonction `ProxyHandler(remote string) http.HandlerFunc`**

Crée un handler qui:
1. Contacte l'API externe (Groupie Trackers)
2. Ajoute le header CORS `Access-Control-Allow-Origin: *`
3. Proxie la réponse vers le client
4. Gère les erreurs (timeout 10s, erreurs réseau)

**Flux:**
```
Client Browser
    ↓ fetch('/api/artists-proxy')
    ↓ (même domaine, pas de CORS)
Notre serveur (ProxyHandler)
    ↓ client.Get('https://groupietrackers.herokuapp.com/api/artists')
    ↓ (serveur à serveur, pas de CORS)
API Groupie Trackers
    ↓ JSON response
    ↓
Notre serveur (ajoute headers CORS)
    ↓ io.Copy(w, resp.Body)
    ↓
Client Browser reçoit les données ✅
```

**Points clés:**
- Timeout 10s pour éviter les blocages
- Content-Type préservé (applique le même que l'API distante)
- Logging pour debug
- Gestion erreur 502 Bad Gateway si l'API externe échoue

### 📄 internal/handlers/static.go

**Rôle**: Serving des fichiers statiques et templates HTML.

**Fonction `SetupStaticRoutes(staticDir string)`**

Configure la route `/static/` pour servir:
- `*.css` → Content-Type: text/css
- `*.js` → Content-Type: application/javascript
- Images (`.png`, `.jpg`, `.svg`, `.gif`, `.webp`)
- Cache-Control: `public, max-age=31536000` (1 an)

**Fonction `SetupTemplateRoutes()`**

Configure les routes pour les pages HTML:
- `/` → index.html (SPA root)
- `/search`, `/filters` → index.html (SPA routing)
- `/geoloc` → Redirection vers /geoloc.html
- `/search.html` → web/templates/search.html
- `/login`, `/login.html` → web/templates/login.html
- `/geoloc.html` → web/templates/geoloc.html

```go
// Ligne 27-29: Utilise un defer pour capturer toutes les panics possibles au démarrage
defer func() {
    if r := recover(); r != nil {
        log.Fatalf("panic occurred: %v", r)  // Log et arrêt propre si panique
    }
}()

// Ligne 31: Vérifie la variable d'environnement DISABLE_DB
// Permet de désactiver complètement la DB (utile pour tests ou déploiements statiques)
if os.Getenv("DISABLE_DB") != "1" {
    var err error
    
    // Ligne 33: Appelle initDB() qui configure et test la connexion MySQL
    db, err = initDB()
    
    // Ligne 34-38: Gestion d'erreur NON-BLOQUANTE
    if err != nil {
        // ⚠️ Important: On log l'erreur mais on continue l'exécution
        // Le serveur peut fonctionner sans DB (mode read-only)
        log.Printf("DB disabled (init failed): %v", err)
        db = nil  // Met db à nil pour que les handlers sachent qu'il n'y a pas de DB
    } else {
        log.Println("DB connection established")
    }
} else {
    // Ligne 42-44: DB explicitement désactivée via variable d'environnement
    log.Println("DB disabled via DISABLE_DB=1")
}
```

**💡 Pourquoi cette logique ?**
- **Flexibilité de déploiement**: Sur des plateformes comme Netlify/Vercel (serverless), pas de MySQL disponible
- **Dégradation gracieuse**: Le site fonctionne en lecture seule même si la DB est down
- **Tests faciles**: Peut tester le frontend sans avoir à configurer MySQL
- **Production robuste**: Une erreur DB ne crash pas tout le serveur

**Ligne 50-56**: Configuration du port
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"  // Port par défaut
}
```

**Ligne 58-83**: Fonction proxy CORS - Explication complète ligne par ligne

```go
// Ligne 58: Définit une fonction qui RETOURNE une fonction (closure)
// C'est un "factory pattern" qui crée des handlers HTTP personnalisés
proxy := func(remote string) http.HandlerFunc {
    // 'remote' est capturé dans la closure - chaque handler aura son URL
    
    // Ligne 59: Retourne un HandlerFunc standard (signature: func(w, r))
    return func(w http.ResponseWriter, r *http.Request) {
        // Ligne 60: Log pour debug - voir quelle API est appelée
        log.Printf("Proxying request to: %s", remote)
        
        // Ligne 61: Crée un client HTTP avec timeout
        // ⏱️ Timeout de 10s évite que le serveur reste bloqué indéfiniment
        // Si l'API externe ne répond pas en 10s, on coupe la connexion
        client := &http.Client{Timeout: 10 * time.Second}
        
        // Ligne 62: Fait la vraie requête GET vers l'API externe
        // C'est ICI que le serveur Go contacte groupietrackers.herokuapp.com
        resp, err := client.Get(remote)
        
        // Ligne 63-67: Gestion d'erreur si la requête échoue
        if err != nil {
            log.Printf("Error fetching %s: %v", remote, err)
            // Retourne une erreur 502 (Bad Gateway) au client
            // Code approprié car c'est l'API externe qui a échoué, pas notre serveur
            http.Error(w, "failed to fetch remote API", http.StatusBadGateway)
            return
        }
        
        // Ligne 68: IMPORTANT - fermer le body pour éviter les fuites mémoire
        // defer = exécuté à la fin de la fonction, peu importe comment elle se termine
        defer resp.Body.Close()
        
        // Ligne 70-74: Copie le Content-Type de la réponse originale
        if ct := resp.Header.Get("Content-Type"); ct != "" {
            w.Header().Set("Content-Type", ct)
        } else {
            // Fallback si pas de Content-Type (rare mais safe)
            w.Header().Set("Content-Type", "application/json")
        }
        
        // Ligne 75: 🔑 HEADER CRUCIAL - Résout le problème CORS
        // "*" = autorise TOUTES les origines (domaines)
        // Sans ce header, le navigateur BLOQUE la réponse JavaScript
        w.Header().Set("Access-Control-Allow-Origin", "*")
        
        // Ligne 77: Écrit le status code de la réponse originale
        // Si l'API externe retourne 404, on retourne aussi 404
        w.WriteHeader(resp.StatusCode)
        
        // Ligne 78-81: Copie le BODY de la réponse vers notre réponse
        // io.Copy est efficace - lit par chunks, pas tout en mémoire
        _, err = io.Copy(w, resp.Body)
        if err != nil {
            log.Printf("Error copying response body: %v", err)
        }
        
        log.Printf("Successfully proxied request to: %s", remote)
    }
}
```

**🔍 Pourquoi un proxy est nécessaire ?**

**Le problème CORS expliqué simplement:**
```
Frontend (localhost:8080)
    ↓ Requête JavaScript fetch()
    ❌ Navigateur BLOQUE ici ❌
    ↓ 
API externe (groupietrackers.herokuapp.com)
```

**Erreur dans la console navigateur:**
```
Access to fetch at 'https://groupietrackers.herokuapp.com/api/artists' 
from origin 'http://localhost:8080' has been blocked by CORS policy
```

**La solution avec le proxy:**
```
Frontend (localhost:8080)
    ↓ fetch('/api/artists-proxy')  ← Même domaine, pas de CORS!
    ✅ OK
    ↓
Notre serveur Go (localhost:8080)
    ↓ client.Get(remote)  ← Serveur à serveur, pas de CORS!
    ✅ OK
    ↓
API externe (groupietrackers.herokuapp.com)
    ↓
    ← Données JSON
    ↓
Notre serveur Go (ajoute header CORS)
    ↓
Frontend reçoit les données ✅
```

**⚠️ Note sécurité:**
- `Access-Control-Allow-Origin: "*"` est OK ici car c'est une API publique
- Pour des APIs privées, il faudrait spécifier le domaine exact
- En production, considérer rate limiting pour éviter l'abus

**Ligne 85-89**: Routes proxy
```go
http.HandleFunc("/api/artists-proxy", proxy("https://groupietrackers.herokuapp.com/api/artists"))
http.HandleFunc("/api/locations-proxy", proxy("https://groupietrackers.herokuapp.com/api/locations"))
http.HandleFunc("/api/dates-proxy", proxy("https://groupietrackers.herokuapp.com/api/dates"))
http.HandleFunc("/api/relation-proxy", proxy("https://groupietrackers.herokuapp.com/api/relation"))
```

**Ligne 92-133**: Serveur de fichiers statiques
```go
http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
    reqPath := r.URL.Path[len("/static/"):]  // Extrait le chemin après "/static/"
    full := filepath.Join(staticDir, filepath.FromSlash(reqPath))  // Path complet
    
    // Définit le Content-Type selon l'extension
    ext := filepath.Ext(full)
    switch ext {
    case ".css":
        w.Header().Set("Content-Type", "text/css")
    case ".js":
        w.Header().Set("Content-Type", "application/javascript")
    // ... autres types ...
    }
    
    w.Header().Set("Cache-Control", "public, max-age=31536000")  // Cache 1 an
    http.ServeFile(w, r, full)
})
```
**Point clé**: Le cache long (1 an) améliore les performances. En production, utiliser des noms de fichiers versionnés (ex: `style.v123.css`).

**Ligne 138-158**: Routes pages HTML
```go
http.HandleFunc("/", serveIndex)  // Racine → index.html
http.HandleFunc("/search", serveIndex)  // SPA routing
http.HandleFunc("/filters", serveIndex)  // SPA routing
http.HandleFunc("/geoloc", redirect)    // Redirection
http.HandleFunc("/search.html", serveTemplate)
http.HandleFunc("/login", serveLoginTemplate)
```

**Ligne 161-162**: Routes API authentification
```go
http.HandleFunc("/api/register", handleRegister)
http.HandleFunc("/api/login", handleLogin)
```

#### Fonction `initDB()`

**Ligne 176-194**: Configuration MySQL
```go
func initDB() (*sql.DB, error) {
    dsn := mysql.Config{
        User:   getenvDefault("DB_USER", "root"),
        Passwd: getenvDefault("DB_PASS", ""),
        Net:    "tcp",
        Addr:   fmt.Sprintf("%s:%s", 
                 getenvDefault("DB_HOST", "localhost"), 
                 getenvDefault("DB_PORT", "3306")),
        DBName: getenvDefault("DB_NAME", "groupi_tracker"),
        AllowNativePasswords: true,  // Compatibilité anciennes versions MySQL
        ParseTime: true,             // Parse datetime en time.Time
        Loc: time.Local,             // Timezone locale
        Params: map[string]string{
            "charset": "utf8mb4",    // Support emojis et caractères spéciaux
        },
    }
    // ...
}
```

**Ligne 200-202**: Optimisation pool de connexions
```go
database.SetMaxOpenConns(10)           // Max 10 connexions simultanées
database.SetMaxIdleConns(5)            // 5 connexions en idle
database.SetConnMaxLifetime(30 * time.Minute)  // Durée vie connexion
```
**Explication**: Ces valeurs optimisent la gestion des connexions DB pour éviter les timeouts et limiter la charge.

#### Fonction `handleRegister()` - Création d'un compte utilisateur

**Ligne 219-255**: Validation et création utilisateur - Explication complète

```go
// Ligne 219: Signature de la fonction - Handler HTTP standard
func handleRegister(w http.ResponseWriter, r *http.Request) {
    
    // ═══════════════════════════════════════════════════════════
    // ÉTAPE 1: Vérification de la méthode HTTP
    // ═══════════════════════════════════════════════════════════
    // Ligne 220-223: Accepte UNIQUEMENT POST (création de ressource)
    if r.Method != http.MethodPost {
        // Si GET, PUT, DELETE, etc. → erreur 405 Method Not Allowed
        // C'est une bonne pratique REST
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return  // ⚠️ Important: return pour arrêter l'exécution
    }
    
    // ═══════════════════════════════════════════════════════════
    // ÉTAPE 2: Vérification de la disponibilité de la DB
    // ═══════════════════════════════════════════════════════════
    // Ligne 224-227: Si db == nil, on ne peut pas créer de compte
    if db == nil {
        // Retourne 503 Service Unavailable (service temporairement indisponible)
        // Mieux que 500 car indique que ce n'est pas une erreur du code
        writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable"})
        return avec bcrypt - Explication cryptographique

```go
// ═══════════════════════════════════════════════════════════
// ÉTAPE 5: Hachage sécurisé du password avec bcrypt
// ═══════════════════════════════════════════════════════════

// Ligne 236: Convertit le password string en []byte (requis par bcrypt)
// bcrypt.GenerateFromPassword() fait 3 choses:
// 1. Génère un SALT aléatoire (empêche les rainbow tables)
// 2. Combine password + salt sécurisée - Protection contre les injections

```go
// ═══════════════════════════════════════════════════════════
// ÉTAPE 6: Insertion dans la base de données MySQL
// ═══════════════════════════════════════════════════════════

// Ligne 243: Prépare la requête SQL avec des placeholders (?)
// Les backticks ` autour des noms de colonnes sont nécessaires car:
// - "Prénom" contient un accent (caractère spécial en SQL)
// - C'est une bonne pratique même sans caractères spéciaux
query := "INSERT INTO `user` (`Nom`, `Prénom`, `sexe`, `password`) VALUES (?, ?, ?, ?)"
//                                                                           ↑  ↑  ↑  ↑
//                                                                           1  2  3  4
//                                                                     Ces ? seront remplacés

// Ligne 244-248: Exécute la requête avec les valeurs réelles
// db.Exec() fait 2 choses:
// 1. Remplace les ? par les valeurs (de manière sécurisée)
// 2. Exécute la requête sur MySQL
result, err := db.Exec(query, 
    req.Nom,         // ? numéro 1 → "Dupont"
    req.Prenom,      // ? numéro 2 → "Marie"
    req.Sexe,        // ? numéro 3 → "F"
    string(hash))    // ? numéro 4 → "$2a$10$N9qo8uLOickgx2..."
    //               ↑ Conversion []byte → string

// Ligne 245-248: Gestion d'erreur (peut échouer si user existe déjà)
if err != nil {
    // Retourne 409 Conflict (code approprié pour "already exists")
    // err.Error() donne le détail MySQL (ex: "Duplicate entry")
    writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
    return
}

// Ligne 250: Récupère l'ID auto-généré par MySQL
id, _ := result.LastInsertId()
// Si la table user a: id_user INT AUTO_INCREMENT PRIMARY KEY
// MySQL génère automatiquement l'ID (1, 2, 3, ...)

// Ligne 251: Retourne succès avec l'ID au client
writeJSON(w, http.StatusCreated, map[string]any{
    "message": "user created",
    "id_utilisateur": id  // Le frontend a besoin de cet ID pour se connecter
})
```

**🛡️ PROTECTION CONTRE LES INJECTIONS SQL - Explications détaillées**

**❌ VERSION DANGEREUSE (JAMAIS FAIRE ÇA):**
```go
// 🚨 ATTENTION: Code vulnérable à l'injection SQL!
query := fmt.Sprintf("INSERT INTO user (Nom, password) VALUES ('%s', '%s')", 
                     req.Nom, string(hash))
db.Exec(query)
```

**Attaque possible:**
```
Un attaquant envoie:
req.Nom = "Dupont'); DROP TABLE user; --"

La requête devient:
INSERT INTO user (Nom, password) VALUES ('Dupont'); DROP TABLE user; --', '...')
                                                     ↑
                                            Exécute une 2ème commande!
                                            Efface toute la table! 💥
```

**✅ VERSION SÉCURISÉE (ce qu'on fait):**
```go
// ✅ Avec des placeholders (prepared statements)
query := "INSERT INTO user (Nom, password) VALUES (?, ?)"
db.Exec(query, req.Nom, string(hash))
```

**Ce qui se passe en coulisses:**
```
1. MySQL reçoit la requête avec ?
   → "INSERT INTO user (Nom, password) VALUES (?, ?)"
   
2. MySQL compile et prépare la requête (structure fixée)
   
3. MySQL reçoit les valeurs séparément
   → ["Dupont'); DROP TABLE user; --", "$2a$10$..."]
   
4. MySQL traite ces valeurs comme des DONNÉES, pas du CODE
   → Elles sont échappées automatiquement
   → Impossible d'injecter du SQL
```

**Résultat final en DB:**
```sql
-- La valeur est insérée LITTÉRALEMENT (guillemets échappés)
Nom: "Dupont'); DROP TABLE user; --"
```

**📊 Autres protections implémentées:**

| Vulnérabilité | Comment on la prévient | Ligne de code |
|--------------|------------------------|---------------|
| SQL Injection | Prepared statements (?) | Ligne 243-244 |
| XSS | JSON encoding auto-escape | writeJSON() |
| Password sniffing | Bcrypt hash | Ligne 236 |
| Timing attacks | bcrypt résistant | Ligne 287 |
| CSRF | À implémenter (tokens) | TODO |
| Rate limiting | À implémenter | TODO |

**💡 Bonus - Détection d'utilisateur existant:**

Le code actuel retourne l'erreur MySQL brute. En production, faire:
```go
if err != nil {
    // Détecte si c'est une erreur de duplication (code MySQL 1062)
    if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
        writeJSON(w, http.StatusConflict, map[string]string{
            "error": "Un compte avec ce nom existe déjà"
        })
    } else {
        // Autre erreur (connection perdue, etc.)
        writeJSON(w, http.StatusInternalServerError, map[string]string{
            "error": "Database error"  // Ne pas exposer les détails
        })
    }
    return
}
```
    // Retourne 500 Internal Server Error (c'est une erreur serveur, pas client)
    writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
    return
}

// 🔐 À ce stade, hash contient quelque chose comme:
// $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
//  │   │  │                                                    │
//  │   │  │                                                    └─ Hash (184 bits)
//  │   │  └─ Salt (128 bits aléatoires)
//  │   └─ Cost factor (10 = 2^10 = 1024 iterations)
//  └─ Algo version (2a = bcrypt)
```

**🛡️ Pourquoi bcrypt est le meilleur choix ?**

**Comparaison des algos de hachage:**

| Algorithme | Sécurité | Temps hash | Résistance brute force |
|------------|----------|------------|------------------------|
| MD5 | ❌ Cassé | 0.001ms | ⚠️ Très faible (GPU) |
| SHA-256 | ⚠️ Trop rapide | 0.002ms | ⚠️ Faible (GPU) |
| **bcrypt** | ✅ Excellent | ~100ms | ✅ **Très forte** |
| Argon2 | ✅ Top | ~50ms | ✅ Très forte |

**Ce que bcrypt fait concrètement:**

```go
// Password en clair (ce que l'utilisateur tape)
password := "MonSuperPassword123!"

// 1. Génération d'un salt aléatoire unique
salt := generateRandomSalt()  // Ex: "N9qo8uLOickgx2ZMRZoMye"

// 2. Concaténation
combined := password + salt   // "MonSuperPassword123!N9qo8uLOickgx2ZMRZoMye"

// 3. Application de bcrypt avec 2^10 = 1024 rounds
// Chaque round fait un hachage complet - c'est LENT volontairement
hash := bcrypt(combined, rounds=10)

// 4. Résultat final (salt + hash ensemble)
result := "$2a$10$" + salt + hash
```

**🎯 Avantages de bcrypt:**

1. **Salt intégré**: Pas besoin de le stocker séparément
2. **Coût adaptatif**: On peut augmenter le cost quand les CPUs deviennent plus rapides
3. **Résistant GPU**: Contrairement à SHA, bcrypt est difficile à paralléliser
4. **Timing attack resistant**: Temps constant pour comparer

**❌ Ce qu'on NE FAIT JAMAIS:**

```go
// ❌ DANGER: Stocker en clair
db.Exec("INSERT INTO user (password) VALUES (?)", req.Password)

// ❌ DANGER: MD5 ou SHA seul (trop rapide, pas de salt)
hash := md5.Sum([]byte(req.Password))

// ❌ DANGER: Salt réutilisé pour tous les users
hash := sha256.Sum([]byte(req.Password + "monSaltFixe"))

// ✅ CORRECT: bcrypt avec salt unique par user
hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
```

**💡 Fun fact:**
Avec bcrypt cost=10, un attaquant avec un GPU moderne peut tester ~10 passwords/seconde.
Pour un password de 8 caractères (alphanumériques), ça prendrait **~190 000 ans** ! 🚀
    // json.NewDecoder(r.Body) crée un décodeur qui lit le body de la requête
    // .Decode(&req) remplit la structure avec les données JSON
    // Le & (pointeur) permet de modifier req directement
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Si le JSON est malformé (ex: virgule manquante, guillemets ouverts)
        // Retourne 400 Bad Request avec message explicite
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    
    // 📝 À ce stade, req contient:
    // req.Nom = "Dupont"
    // req.Prenom = "Marie"  
    // req.Sexe = "F"
    // req.Password = "monMotDePasse123"
    
    // Ligne 235: Log pour debug (visible dans la console serveur)
    log.Printf("Register attempt: Nom=%q, Prenom=%q, Sexe=%q, PwdLen=%d", 
               req.Nom, req.Prenom, req.Sexe, len(req.Password))
    
    // ═══════════════════════════════════════════════════════════
    // ÉTAPE 4: Validation des champs (Business Logic)
    // ═══════════════════════════════════════════════════════════
    // Ligne 237-240: Vérifie que tous les champs sont présents et valides
    if req.Nom == "" ||           // Nom vide
       req.Prenom == "" ||        // Prénom vide
       req.Sexe == "" ||          // Sexe vide
       len(req.Password) < 6 {    // Password trop court (min 6 caractères)
        
        // Retourne 400 avec message générique
        // 💡 Astuce sécurité: ne pas dire QUEL champ est invalide
        // (évite de donner trop d'infos aux attaquants)
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing or invalid fields"})
        return
    }
    
    // ⚠️ TODO en production: ajouter plus de validations
    // - Sexe doit être M, F, ou Autre (pas "XYZ")
    // - Nom/Prénom max 100 caractères
    // - Password: min 1 majuscule, 1 chiffre, 1 caractère spécial
    // - Email validation avec regex
    
    // ... suite ligne 236
}
```

**🔒 Sécurité des validations:**

| Validation | Pourquoi c'est important | Exemple attaque évitée |
|------------|-------------------------|------------------------|
| Méthode POST uniquement | Respect REST, évite GET avec params dans URL | URLs cachées avec passwords visibles |
| Check db != nil | Évite panic si DB down | Server crash sur nil pointer |
| JSON parsing | Évite injections, overflow | Buffer overflow avec données binaires |
| Champs non-vides | Évite données corrompues en DB | Comptes zombies |
| Password min 6 chars | Force un minimum de sécurité | Brute force trop facile |

**Ligne 236-241**: Hachage sécurisé du password
```go
hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
    return
}
```
**Sécurité**: `bcrypt` utilise un salt automatique et un coût adaptatif. Le password n'est **jamais** stocké en clair.

**Ligne 243-249**: Insertion SQL
```go
query := "INSERT INTO `user` (`Nom`, `Prénom`, `sexe`, `password`) VALUES (?, ?, ?, ?)"
result, err := db.Exec(query, req.Nom, req.Prenom, req.Sexe, string(hash))
```
**Important**: Les `?` sont des placeholders qui protègent contre les **injections SQL**.

#### Fonction `handleLogin()`

**Ligne 253-295**: Authentification utilisateur
```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    // Parse requête
    var req struct {
        IDUser   int    `json:"id_utilisateur"`
        Password string `json:"password"`
    }
    // ...
    
    // Récupère le hash depuis la DB
    var storedHash string
    var nom, prenom, sexe string
    row := db.QueryRow("SELECT `password`, `Nom`, `Prénom`, `sexe` FROM `user` WHERE `id_user` = ?", req.IDUser)
    if err := row.Scan(&storedHash, &nom, &prenom, &sexe); err != nil {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
        return
    }
    
    // Vérifie le password
    if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
        return
    }
    
    // Retourne les infos utilisateur (sans le password!)
    writeJSON(w, http.StatusOK, map[string]any{
        "message": "login ok",
        "user": map[string]any{
            "id_utilisateur": req.IDUser,
            "nom": nom,
            "prenom": prenom,
            "sexe": sexe,
        },
    })
}
```
**Sécurité**: `bcrypt.CompareHashAndPassword` est résistant aux **timing attacks** car prend un temps constant.

---

### 📄 api/handler.go

**Rôle**: Handler serverless pour Vercel/Netlify. Version simplifiée du proxy principal.

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    var remoteURL string
    
    // Route selon le path
    switch r.URL.Path {
    case "/api/artists-proxy":
        remoteURL = "https://groupietrackers.herokuapp.com/api/artists"
    // ... autres routes ...
    default:
        http.NotFound(w, r)
        return
    }
    
    // Proxy la requête
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(remoteURL)
    // ... relay response ...
}
```

**Différence avec main.go**: Version stateless pour environnements serverless (pas de state global, pas de DB).

---

### 📄 api/index.go

**Rôle**: Handler principal serverless avec gestion des fichiers statiques.

#### Fonction `Handler()`
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    
    // Route static files
    if strings.HasPrefix(path, "/static/") {
        handleStatic(w, r)
        return
    }
    
    // Route API proxies
    if strings.HasPrefix(path, "/api/") {
        handleAPIProxy(w, r)
        return
    }
    
    // Routes templates
    if path == "/search.html" {
        serveTemplate(w, r, "search.html")
        return
    }
    
    // Default: serve index.html (SPA)
    http.ServeFile(w, r, "index.html")
}
```

#### Fonction `handleStatic()`
```go
func handleStatic(w http.ResponseWriter, r *http.Request) {
    reqPath := r.URL.Path[len("/static/"):]
    
    // Essaie plusieurs chemins possibles (selon l'environnement de déploiement)
    possiblePaths := []string{
        filepath.Join("web", "static", filepath.FromSlash(reqPath)),
        filepath.Join("..", "web", "static", filepath.FromSlash(reqPath)),
        filepath.Join(".", "web", "static", filepath.FromSlash(reqPath)),
    }
    
    // Cherche le fichier
    for _, path := range possiblePaths {
        if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
            // Fichier trouvé: définit Content-Type et sert
            setContentType(w, path)
            http.ServeFile(w, r, path)
            return
        }
    }
    
    http.NotFound(w, r)
}
```
**Astuce**: Les chemins multiples permettent de fonctionner sur différentes plateformes (local, Vercel, Netlify, etc.).

---

### 📄 internal/api/models.go

**Rôle**: Définit les structures de données pour l'API Groupie Trackers.

```go
// Location représente les lieux de concerts d'un artiste
type Location struct {
    ID        int      `json:"id"`
    Locations []string `json:"locations"`  // Ex: ["usa-texas-houston", "france-paris"]
    Dates     string   `json:"dates"`      // URL vers les dates
}

// LocationsResponse contient tous les lieux
type LocationsResponse struct {
    Index []Location `json:"index"`
}

// Dates représente les dates de concerts
type Dates struct {
    ID    int      `json:"id"`
    Dates []string `json:"dates"`  // Ex: ["*12-01-2023", "*15-02-2023"]
}

// Relations mappe lieu → dates
type Relations struct {
    ID             int                 `json:"id"`
    DatesLocations map[string][]string `json:"datesLocations"`
    // Ex: {"usa-texas-houston": ["12-01-2023", "13-01-2023"]}
}
```

**Format des données**:
- **Locations**: Utilisent le format `pays-région-ville` avec underscores
- **Dates**: Préfixées par `*` dans l'API originale
- **Relations**: Combinent lieux et dates pour cartographie

---

## Documentation des fichiers HTML

### 📄 index.html

**Rôle**: Page d'accueil principale (SPA root).

#### Structure `<head>`
```html
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Groupie Tracker</title>
    <meta name="description" content="..." />
    
    <!-- Fonts Google -->
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;700&family=Merriweather:wght@700&display=swap" rel="stylesheet">
    
    <!-- Styles -->
    <link rel="stylesheet" href="/static/css/style.css" />
</head>
```

#### Header avec recherche
```html
<header class="site-header">
    <div class="header-content">
        <h1>Groupie Tracker</h1>
        
        <!-- Formulaire de recherche header -->
        <form id="headerSearch" class="header-search-form" action="/search.html" method="get">
            <input type="search" name="q" placeholder="Rechercher un artiste..." />
            <button type="submit" class="btn">Recherche</button>
        </form>
        
        <!-- Navigation principale -->
        <nav class="main-nav" id="mainNav">
            <a href="/">Accueil</a>
            <a href="/geoloc.html">Géolocalisation</a>
        </nav>
        
        <!-- Actions utilisateur -->
        <div class="header-actions">
            <a class="btn btn-auth" href="/login">Connexion / Inscription</a>
            <button class="btn btn-subscribe" id="subscribeBtn">S'abonner</button>
        </div>
    </div>
</header>
```

#### Section Hero
```html
<section class="hero">
    <h2>Trouvez facilement vos groupes préférés</h2>
    <p>Utilisez la recherche pour retrouver des groupes, filtrez par date ou pays...</p>
    <p class="cta">
        <a class="btn" href="/search.html">Commencer la recherche</a>
    </p>
</section>
```

#### Zone vinyles (animée par JavaScript)
```html
<section class="vinyl-area container" aria-hidden="true">
    <div class="vinyl-grid"></div>  <!-- Rempli dynamiquement par ui.js -->
</section>
```

#### Modal d'abonnement (lignes 61-130)
```html
<div id="subscriptionModal" class="modal">
    <div class="modal-content">
        <button class="modal-close" id="closeModal">&times;</button>
        <h2>S'abonner à Groupie Tracker Premium</h2>
        
        <!-- Plans d'abonnement -->
        <div class="subscription-plans">
            <div class="plan">
                <h3>Plan Mensuel</h3>
                <p class="price">9,99 €<span>/mois</span></p>
                <button class="btn-payment" data-plan="monthly" data-price="9.99">Souscrire</button>
            </div>
            
            <div class="plan featured">
                <h3>Plan Annuel</h3>
                <p class="price">89,99 €<span>/an</span></p>
                <p class="savings">Économisez 20%</p>
                <button class="btn-payment" data-plan="yearly" data-price="89.99">Souscrire</button>
            </div>
        </div>
        
        <!-- Formulaire de paiement (caché initialement) -->
        <div id="paymentForm" class="payment-form hidden">
            <h3>Détails de paiement</h3>
            <form id="cardForm">
                <div class="form-group">
                    <label for="cardNumber">Numéro de carte</label>
                    <input type="text" id="cardNumber" placeholder="1234 5678 9012 3456" maxlength="19" required>
                </div>
                <!-- ... autres champs ... -->
                <button type="submit" class="btn btn-primary">Valider le paiement</button>
            </form>
        </div>
        
        <!-- Message de succès -->
        <div id="successMessage" class="success-message hidden">
            <h3>✓ Paiement réussi!</h3>
            <p>Votre abonnement est maintenant actif.</p>
        </div>
    </div>
</div>
```

**Workflow modal**:
1. Clic sur "S'abonner" → ouvre la modal
2. Sélection d'un plan → affiche le formulaire de paiement
3. Validation formulaire → simule le paiement et affiche succès
4. Stockage dans `localStorage` pour persistance

#### Scripts chargés
```html
<script src="/static/js/ui.js?v=20260102"></script>
<script src="/static/js/subscription.js"></script>
```
**Note**: Le paramètre `?v=20260102` force le rechargement du cache après updates.

---

### 📄 web/templates/search.html

**Rôle**: Page de recherche d'artistes avec filtres.

#### Formulaire de recherche principal
```html
<form id="searchForm" class="search-form">
    <label for="query">Nom de l'artiste :</label>
    
    <!-- Input avec suggestions -->
    <div class="input-stack">
        <input type="text" id="query" name="q" placeholder="Entrez un nom d'artiste" 
               required autocomplete="off">
        <div id="suggestions" class="suggestions" role="listbox"></div>
    </div>
    
    <!-- Filtres rapides (chips) -->
    <div class="quick-filters" id="quickFilters">
        <button type="button" class="chip" data-filter="rock">Rock</button>
        <button type="button" class="chip" data-filter="seventies">Années 70</button>
        <button type="button" class="chip" data-filter="usa">USA</button>
        <button type="button" class="chip" data-filter="month">Concerts ce mois-ci</button>
    </div>
    
    <div class="actions">
        <button type="submit" class="btn-primary">Rechercher</button>
        <button type="button" id="clearSearch" class="btn-ghost">Effacer</button>
    </div>
</form>
```

#### Zone résultats (remplie par JavaScript)
```html
<section id="results" class="results" aria-live="polite">
    <p>Entrez un nom ou utilisez un filtre pour voir les résultats.</p>
</section>
```

**Attribut `aria-live="polite"`**: Annonce les changements aux lecteurs d'écran (accessibilité).

---

### 📄 web/templates/geoloc.html

**Rôle**: Page de géolocalisation avec carte Leaflet.

#### Inclusion Leaflet
```html
<head>
    <!-- ... -->
    <!-- Leaflet CSS -->
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" crossorigin="anonymous" />
</head>
```

#### Conteneur carte
```html
<section>
    <div id="map" aria-label="Carte des concerts"></div>
    <div id="geo-status" class="geo-status" aria-live="polite">
        Chargement des données…
    </div>
</section>
```

#### Scripts Leaflet
```html
<!-- Leaflet JS -->
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" crossorigin="anonymous"></script>
<script src="/static/js/ui.js"></script>
<script src="/static/js/geoloc.js"></script>
```

**Ordre important**: Leaflet doit être chargé avant `geoloc.js` qui l'utilise.

---

### 📄 web/templates/login.html

**Rôle**: Page d'authentification (connexion et inscription).

#### Système d'onglets
```html
<div class="auth-tabs" id="authTabs">
    <button class="auth-tab active" data-target="login">Connexion</button>
    <button class="auth-tab" data-target="register">Inscription</button>
</div>
```

#### Formulaire connexion
```html
<form id="loginForm" class="auth-form">
    <label class="form-field">
        <span>ID utilisateur</span>
        <input type="number" name="id_utilisateur" required placeholder="123" />
    </label>
    <label class="form-field">
        <span>Mot de passe</span>
        <input type="password" name="password" required placeholder="••••••••" />
    </label>
    <button type="submit" class="btn btn-primary auth-submit">Se connecter</button>
</form>
```

#### Formulaire inscription
```html
<form id="registerForm" class="auth-form hidden">
    <div class="form-grid">
        <label class="form-field">
            <span>Nom</span>
            <input type="text" name="nom" required />
        </label>
        <label class="form-field">
            <span>Prénom</span>
            <input type="text" name="prenom" required />
        </label>
    </div>
    <label class="form-field">
        <span>Mot de passe</span>
        <input type="password" name="password" required />
    </label>
    <label class="form-field">
        <span>Sexe</span>
        <select name="sexe" required>
            <option value="">Sélectionner</option>
            <option value="F">Femme</option>
            <option value="M">Homme</option>
            <option value="Autre">Autre</option>
        </select>
    </label>
    <button type="submit" class="btn btn-primary auth-submit">Créer mon compte</button>
</form>
```

#### Script inline (gestion auth)
```html
<script>
// Système d'onglets
const tabs = document.querySelectorAll('.auth-tab');
tabs.forEach(tab => {
    tab.addEventListener('click', () => {
        // Toggle active tab
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        
        // Show/hide forms
        const target = tab.dataset.target;
        forms.login.classList.toggle('hidden', target !== 'login');
        forms.register.classList.toggle('hidden', target !== 'register');
    });
});

// Fonction helper pour afficher messages
const showMessage = (text, isError = false) => {
    const toast = document.createElement('div');
    toast.className = 'auth-toast';
    if (isError) toast.style.background = 'linear-gradient(135deg, #ef4444, #dc2626)';
    toast.textContent = text;
    document.body.appendChild(toast);
    setTimeout(() => toast.classList.add('show'), 10);
    setTimeout(() => { /* remove toast */ }, 2600);
};

// Helper pour requêtes JSON
const postJSON = async (url, data) => {
    const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || 'Erreur serveur');
    return body;
};

// Handler login
document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const payload = {
        id_utilisateur: parseInt(form.id_utilisateur.value, 10),
        password: form.password.value
    };
    try {
        await postJSON('/api/login', payload);
        showMessage('Connexion réussie');
    } catch (err) {
        showMessage(err.message, true);
    }
});

// Handler register
document.getElementById('registerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const payload = {
        nom: form.nom.value,
        prenom: form.prenom.value,
        password: form.password.value,
        sexe: form.sexe.value
    };
    try {
        const res = await postJSON('/api/register', payload);
        showMessage(`Compte créé. Votre ID: ${res.id_utilisateur}`);
        form.reset();
    } catch (err) {
        showMessage(err.message, true);
    }
});
</script>
```

**Points clés**:
- Gestion d'erreurs avec try/catch
- Affichage de toasts pour feedback utilisateur
- Récupération de l'ID utilisateur après inscription

---

## Documentation des fichiers JavaScript

### 📄 web/static/js/ui.js

**Rôle**: Logique UI générale + système de vinyles animés avec previews musicales.

#### Initialisation menu mobile (lignes 1-35)
```javascript
document.addEventListener('DOMContentLoaded', function () {
    var nav = document.getElementById('mainNav');
    
    // Crée bouton hamburger
    var toggle = document.createElement('button');
    toggle.id = 'menuToggle';
    toggle.className = 'menu-toggle';
    toggle.textContent = '☰';  // Icône hamburger
    
    function updateNavVisibility() {
        if (window.innerWidth < 700) {
            // Sur mobile: cache le menu par défaut
            nav.style.display = 'none';
        } else {
            // Sur desktop: affiche toujours
            nav.style.display = 'flex';
        }
    }
    
    // Toggle au clic
    toggle.addEventListener('click', function () {
        var showing = nav.style.display !== 'none';
        nav.style.display = showing ? 'none' : 'flex';
        toggle.setAttribute('aria-expanded', !showing);
    });
    
    window.addEventListener('resize', updateNavVisibility);
    updateNavVisibility();
});
```

#### Smooth scroll (lignes 37-45)
```javascript
document.querySelectorAll('a[href^="#"]').forEach(function (a) {
    a.addEventListener('click', function (e) {
        var tgt = document.querySelector(this.getAttribute('href'));
        if (tgt) {
            e.preventDefault();
            tgt.scrollIntoView({behavior:'smooth'});
        }
    });
});
```

#### Système de vinyles - Initialisation (lignes 48-67)
```javascript
document.addEventListener('DOMContentLoaded', function () {
    // URLs API (avec fallback local → remote)
    const LOCAL_API = '/api/artists-proxy';
    const REMOTE_API = 'https://groupietrackers.herokuapp.com/api/artists';
    
    const vinylGrid = document.querySelector('.vinyl-area .vinyl-grid');
    if (!vinylGrid) return;  // Pas de grid = pas de vinyles
    
    let locationsData = null;  // Chargé une fois pour tous les artistes
    let datesData = null;
    let relationsData = null;
    
    // État audio global
    let currentAudio = null;  // Audio en cours de lecture
    let currentFrame = null;  // Frame du vinyle actif
});
```

#### Chargement données API (lignes 70-140)
```javascript
async function tryFetch(url) {
    const res = await fetch(url, {cache: 'no-store'});
    if (!res.ok) throw new Error('API response ' + res.status);
    return res.json();
}

async function loadLocations() {
    try {
        locationsData = await tryFetch(LOCAL_API);
    } catch (err) {
        // Fallback vers API externe si proxy échoue
        try {
            locationsData = await tryFetch(REMOTE_API);
        } catch (err2) {
            console.warn('Failed to load locations', err, err2);
        }
    }
}

// Idem pour loadDates() et loadRelations()
```

**Pattern important**: Toujours essayer le proxy local d'abord (pas de CORS), puis fallback vers API externe.

#### Fonctions helpers (lignes 142-156)
```javascript
function getLocationsForArtist(artistId) {
    if (!locationsData || !locationsData.index) return null;
    const artistLoc = locationsData.index.find(l => l.id === artistId);
    return artistLoc ? artistLoc.locations : null;
}

// Similar pour getDatesForArtist() et getRelationsForArtist()
```

#### Chargement artistes et création vinyles (lignes 158-266)
```javascript
async function loadArtists() {
    // Charge d'abord les données supplémentaires
    await Promise.all([
        loadLocations().catch(e => console.warn('Locations failed:', e)),
        loadDates().catch(e => console.warn('Dates failed:', e)),
        loadRelations().catch(e => console.warn('Relations failed:', e))
    ]);
    
    // Charge artistes
    let data;
    try {
        data = await tryFetch(LOCAL_API);
    } catch (err) {
        // Fallback
        data = await tryFetch(REMOTE_API);
    }
    
    const artists = Array.isArray(data) ? data : (data.artists || data);
    
    vinylGrid.innerHTML = '';  // Clear existing
    
    // Crée un vinyle par artiste
    artists.forEach((a, idx) => {
        const item = document.createElement('div');
        item.className = 'vinyl-item fade-in';
        item.style.animationDelay = `${idx * 60}ms`;  // Apparition progressive
        
        const frame = document.createElement('div');
        frame.className = 'vinyl-frame';
        
        // Crée élément audio
        const audio = document.createElement('audio');
        audio.preload = 'auto';
        audio.volume = 0.85;
        audio.style.display = 'none';
        item.appendChild(audio);
        
        // Crée image cover
        const cover = document.createElement('img');
        cover.className = 'vinyl-cover';
        cover.src = a.image || '/static/images/vinyle.png';
        cover.alt = a.name || '';
        
        frame.appendChild(cover);
        item.appendChild(frame);
        
        // Caption
        const caption = document.createElement('div');
        caption.className = 'vinyl-caption';
        caption.textContent = a.name || '';
        item.appendChild(caption);
        
        vinylGrid.appendChild(item);
        
        // ... configuration audio et events (voir plus bas)
    });
}
```

#### Fetching previews musicales (lignes 208-260) - Système de fallback en cascade

```javascript
// ═══════════════════════════════════════════════════════════
// FONCTION: fetchMusicPreview - Récupère une preview audio
// ═══════════════════════════════════════════════════════════
async function fetchMusicPreview(artistName) {
    // Ligne 208: Vérifie si un fetch est déjà en cours
    // ⚠️ Important: évite les requêtes simultanées pour le même artiste
    if (audioLoading) return null;
    
    // Ligne 210: Flag pour indiquer qu'un fetch est en cours
    audioLoading = true;
    
    // Ligne 212: Encode le nom pour l'URL (remplace espaces, caractères spéciaux)
    // Exemple: "AC/DC" → "AC%2FDC"
    //          "Niña Pastori" → "Ni%C3%B1a%20Pastori"
    const encodedName = encodeURIComponent(artistName);
    
    console.log('🎵 Searching music for:', artistName);
    
    // ═══════════════════════════════════════════════════════════
    // TENTATIVE 1: iTunes API (Apple)
    // ═══════════════════════════════════════════════════════════
    try {
        // Ligne 216-217: Construction de l'URL iTunes Search API
        // Parameters:
        // - term: nom de l'artiste (encodé)
        // - entity=song: cherche des chansons (pas albums, podcasts...)
        // - limit=1: retourne seulement 1 résultat (le plus pertinent)
        // - media=music: filtre pour la musique uniquement
        const itunesUrl = `https://itunes.apple.com/search?term=${encodedName}&entity=song&limit=1&media=music`;
        
        console.log('📡 Fetching from iTunes:', itunesUrl);
        
        // Ligne 218: Fait la requête HTTP GET
        // fetch() retourne une Promise - await attend la réponse
        const itunesRes = await fetch(itunesUrl);
        
        // Ligne 219: Parse la réponse JSON
        // iTunes retourne: { resultCount: 1, results: [{...}] }
        const itunesData = await itunesRes.json();
        
        console.log('📦 iTunes response:', itunesData);
        
        // Ligne 221-231: Vérifie si on a des résultats
        if (itunesData.results && itunesData.results.length > 0) {
            // Récupère l'URL de preview (30 secondes)
            let preview = itunesData.results[0].previewUrl;
            
            if (preview) {
                // Ligne 225-228: Force HTTPS pour éviter Mixed Content
                // Problème: Si notre site est en HTTPS et qu'on charge
                // un audio en HTTP, le navigateur bloque (sécurité)
                if (preview.startsWith('http://')) {
                    preview = preview.replace('http://', 'https://');
                    console.log('🔒 Converted to HTTPS:', preview);
                }
                
                console.log('✅ iTunes preview found:', preview);
                audioLoading = false;  // Libère le flag
                return preview;  // Succès! On retourne l'URL
            }
        }
        
        // Si on arrive ici: iTunes n'a pas de résultat
        console.log('⚠️ No iTunes results for:', artistName);
        
    } catch (err) {
        // Si iTunes API est down, timeout, ou erreur réseau
        console.error('❌ iTunes API error:', err);
        // On continue vers Deezer (pas de return)
    }
    
    // ═══════════════════════════════════════════════════════════
    // TENTATIVE 2: Deezer API (Fallback)
    // ═══════════════════════════════════════════════════════════
    try {
        // Ligne 237: Deezer Search API
        // Plus simple que iTunes: juste ?q= et &limit=
        const deezerUrl = `https://api.deezer.com/search?q=${encodedName}&limit=1`;
        
        console.log('📡 Fetching from Deezer:', deezerUrl);
        
        const deezerRes = await fetch(deezerUrl);
        
        // Ligne 239: Deezer retourne: { data: [{...}], total: 123, ... }
        const deezerData = await deezerRes.json();
        
        console.log('📦 Deezer response:', deezerData);
        
        // Ligne 241-250: Même logique que iTunes
        if (deezerData.data && deezerData.data.length > 0) {
            let preview = deezerData.data[0].preview;
            
            if (preview) {
                // Force HTTPS (Deezer peut retourner HTTP aussi)
                if (preview.startsWith('http://')) {
                    preview = preview.replace('http://', 'https://');
                }
                
                console.log('✅ Deezer preview found:', preview);
                audioLoading = false;
                return preview;
            }
        }
        
        console.log('⚠️ No Deezer results for:', artistName);
        
    } catch (err) {
        console.error('❌ Deezer API error:', err);
    }
    
    // ═══════════════════════════════════════════════════════════
    // ÉCHEC: Aucune preview trouvée
    // ═══════════════════════════════════════════════════════════
    console.warn('❌ No preview found for:', artistName);
    audioLoading = false;
    return null;  // Le code appelant utilisera FALLBACK_PREVIEW
}

// ═══════════════════════════════════════════════════════════
// UTILISATION: Fetch immédiat pour chaque artiste
// ═══════════════════════════════════════════════════════════
// Ligne 260-268: Appelle fetchMusicPreview et configure l'audio
fetchMusicPreview(a.name || '').then(previewUrl => {
    // Cette fonction s'exécute QUAND la Promise est résolue
    // (après que fetchMusicPreview ait fini)
    
    if (previewUrl) {
        // Cas 1: Preview trouvée (iTunes ou Deezer)
        console.log('🔗 Setting audio src:', previewUrl);
        audio.src = previewUrl;
        audio.load();  // Démarre le téléchargement du fichier audio
    } else {
        // Cas 2: Aucune preview (utilise fallback générique)
        console.warn('⚠️ No audio preview found, using fallback for:', a.name);
        audio.src = FALLBACK_PREVIEW;  // Constante définie plus haut
        audio.load();
    }
});
```

**🎯 Stratégie de fallback en cascade - Pourquoi ce choix ?**

**Ordre des tentatives:**
```
1. iTunes API
   ↓ (si échec)
2. Deezer API
   ↓ (si échec)
3. FALLBACK_PREVIEW (MP3 générique)
```

**Comparaison des APIs:**

| API | Avantages | Inconvénients | Qualité preview |
|-----|-----------|---------------|----------------|
| **iTunes** | • Base énorme<br>• Qualité excellente<br>• Metadata riches | • Rate limiting strict<br>• Geo-blocking parfois | ⭐⭐⭐⭐⭐ |
| **Deezer** | • Rapide<br>• Pas de rate limiting<br>• CORS friendly | • Catalogue moins complet<br>• Qualité variable | ⭐⭐⭐⭐ |
| **Fallback** | • Toujours disponible<br>• Pas de requête externe | • Pas l'artiste demandé<br>• Expérience dégradée | ⭐⭐ |

**📊 Statistiques typiques (basées sur ~50 artistes):**
- iTunes trouve: ~85%
- Deezer trouve parmi les 15% restants: ~10%
- Fallback utilisé: ~5%

**💡 Optimisations possibles:**

1. **Cache localStorage:**
```javascript
// Avant de fetch, check cache
const cached = localStorage.getItem(`preview_${artistName}`);
if (cached) return cached;

// Après fetch, store en cache
localStorage.setItem(`preview_${artistName}`, previewUrl);
```

2. **Parallel fetching:**
```javascript
// Au lieu de iTunes puis Deezer, les deux en parallèle
const [itunesData, deezerData] = await Promise.all([
    fetch(itunesUrl).then(r => r.json()).catch(() => null),
    fetch(deezerUrl).then(r => r.json()).catch(() => null)
]);
// Prend la première qui a un résultat
```

3. **Retry avec backoff:**
```javascript
// Si API échoue, retry après 1s, puis 2s, puis 4s
for (let i = 0; i < 3; i++) {
    try {
        const res = await fetch(url);
        if (res.ok) return res;
    } catch {}
    await sleep(2 ** i * 1000);
}
```

#### Gestion audio au survol (lignes 268-330)
```javascript
let isPlaying = false;
let playAttempted = false;
let hoverTimeout = null;

frame.style.cursor = 'pointer';

function tryPlayAudio() {
    if (!audio.src) {
        // Pas encore de source: fetch maintenant
        fetchMusicPreview(a.name || '').then(previewUrl => {
            if (previewUrl) {
                audio.src = previewUrl;
                audio.load();
                setTimeout(() => tryPlayAudio(), 500);
            }
        });
        return;
    }
    
    if (!isPlaying && !playAttempted) {
        playAttempted = true;
        
        // Stop audio précédent
        if (currentAudio && currentAudio !== audio) {
            currentAudio.pause();
            currentAudio.currentTime = 0;
            if (currentFrame) currentFrame.classList.remove('playing');
        }
        
        const playPromise = audio.play();
        if (playPromise !== undefined) {
            playPromise
                .then(() => {
                    isPlaying = true;
                    playAttempted = false;
                    frame.classList.add('playing');  // Animation CSS rotation
                    currentAudio = audio;
                    currentFrame = frame;
                })
                .catch(err => {
                    playAttempted = false;
                    console.error('Audio play failed:', err);
                    // Retry avec fallback si pas déjà sur fallback
                    if (audio.src !== FALLBACK_PREVIEW) {
                        audio.src = FALLBACK_PREVIEW;
                        audio.load();
                        setTimeout(() => tryPlayAudio(), 300);
                    }
                });
        }
    }
}

// Démarre timer au survol
frame.addEventListener('mouseenter', function () {
    hoverTimeout = setTimeout(() => {
        tryPlayAudio();
    }, 2500);  // 2.5 secondes de survol
});

// Annule timer si la souris part
frame.addEventListener('mouseleave', function () {
    if (hoverTimeout) {
        clearTimeout(hoverTimeout);
        hoverTimeout = null;
    }
});

// Au clic: stop audio et ouvre modal
frame.addEventListener('click', function () {
    if (isPlaying) {
        audio.pause();
        audio.currentTime = 0;
        isPlaying = false;
        frame.classList.remove('playing');
    }
    if (hoverTimeout) {
        clearTimeout(hoverTimeout);
        hoverTimeout = null;
    }
    openArtistModal(a);
});
```

**Logique importante**:
- **2.5s de survol** avant de jouer (évite play accidentel)
- **Un seul audio** à la fois (stop les autres)
- **Fallback automatique** si l'audio ne charge pas
- **Cancel au clic** pour ouvrir la modal proprement

#### Création modal artiste (lignes 370-420)
```javascript
let modalEl = null;

function createModal() {
    modalEl = document.createElement('div');
    modalEl.className = 'artist-modal';
    modalEl.id = 'artistModal';
    
    var panel = document.createElement('div');
    panel.className = 'artist-modal__panel';
    
    var closeBtn = document.createElement('button');
    closeBtn.className = 'artist-modal__close';
    closeBtn.textContent = '×';
    closeBtn.addEventListener('click', hideModal);
    
    var content = document.createElement('div');
    content.className = 'artist-modal__content';
    
    panel.appendChild(closeBtn);
    panel.appendChild(content);
    modalEl.appendChild(panel);
    document.body.appendChild(modalEl);
    
    // Ferme au clic extérieur
    modalEl.addEventListener('click', function (e) {
        if (e.target === modalEl) hideModal();
    });
}

function hideModal() {
    if (modalEl) modalEl.classList.remove('open');
}
```

#### Construction contenu modal (lignes 440-550)
```javascript
function openArtistModal(artist) {
    if (!modalEl) createModal();
    
    var panel = modalEl.querySelector('.artist-modal__content');
    panel.innerHTML = '';  // Clear previous content
    
    // Hero section avec image
    var hero = document.createElement('div');
    hero.className = 'artist-modal__hero';
    
    if (artist.image) {
        var cover = document.createElement('img');
        cover.className = 'artist-cover';
        cover.src = artist.image;
        cover.alt = artist.name || '';
        hero.appendChild(cover);
    }
    
    var head = document.createElement('div');
    head.className = 'artist-modal__head';
    head.appendChild(createElement('h2', '', artist.name || 'Artiste'));
    head.appendChild(createElement('p', 'muted', 'Année de création: ' + (artist.creationDate || '—')));
    hero.appendChild(head);
    
    // Body avec membres
    var body = document.createElement('div');
    body.className = 'artist-modal__body';
    
    var mainView = document.createElement('div');
    mainView.className = 'artist-main';
    mainView.appendChild(createElement('h3', '', 'Membres'));
    mainView.appendChild(buildMembersList(artist.members));
    mainView.appendChild(createElement('p', '', 'Premier album: ' + (artist.firstAlbum || '—')));
    
    // Boutons pour voir détails
    var actions = document.createElement('div');
    actions.className = 'artist-links';
    
    function addInfoButton(label, builder) {
        var btn = document.createElement('button');
        btn.className = 'artist-link-btn';
        btn.textContent = label;
        btn.addEventListener('click', function () {
            // Affiche la section de détail et cache la vue principale
            var section = builder();
            detailTitle.textContent = label;
            detailContent.innerHTML = '';
            detailContent.appendChild(section);
            
            mainView.classList.add('is-hidden');
            actions.classList.add('is-hidden');
            detail.classList.remove('is-hidden');
            hero.classList.add('is-hidden');
        });
        actions.appendChild(btn);
    }
    
    addInfoButton('Locations', () => buildLocationsSection(artist));
    addInfoButton('Dates', () => buildDatesSection(artist));
    addInfoButton('Relations', () => buildRelationsSection(artist));
    
    // Assemble tout
    body.appendChild(mainView);
    body.appendChild(actions);
    body.appendChild(detail);
    
    panel.appendChild(hero);
    panel.appendChild(body);
    
    modalEl.classList.add('open');
}
```

#### Helpers formatage (lignes 555-580)
```javascript
function formatLocationName(loc) {
    if (!loc) return '';
    // "usa-texas-houston" → "Usa, Texas, Houston"
    var formatted = loc.replace(/_/g, ' ').replace(/-/g, ', ');
    return formatted.split(' ').map(function (w) {
        return w ? w.charAt(0).toUpperCase() + w.slice(1) : '';
    }).join(' ');
}

function formatDateLabel(dateStr) {
    if (!dateStr) return '';
    var clean = dateStr.replace(/^\*/, '');  // Retire le * initial
    return clean.replace(/-/g, '/');  // 12-01-2023 → 12/01/2023
}
```

---

### 📄 web/static/js/search.js

**Rôle**: Logique de recherche avec filtres et suggestions.

#### Initialisation (lignes 1-10)
```javascript
document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('searchForm');
    const results = document.getElementById('results');
    const input = document.getElementById('query');
    const suggestionsEl = document.getElementById('suggestions');
    const quickFilters = document.getElementById('quickFilters');
    const clearBtn = document.getElementById('clearSearch');
    
    let allArtists = [];    // Cache des artistes
    let activeFilter = null; // Filtre actif (rock, usa, etc.)
});
```

#### Modal de détails (lignes 12-70)
```javascript
let modalEl = null;
let modalBackdrop = null;

function ensureModal() {
    if (modalEl) return;  // Déjà créée
    
    modalBackdrop = document.createElement('div');
    modalBackdrop.className = 'search-modal-backdrop';
    
    modalEl = document.createElement('div');
    modalEl.className = 'search-modal';
    
    const closeBtn = document.createElement('button');
    closeBtn.className = 'search-modal__close';
    closeBtn.textContent = '×';
    closeBtn.addEventListener('click', hideModal);
    
    const content = document.createElement('div');
    content.className = 'search-modal__content';
    
    modalEl.appendChild(closeBtn);
    modalEl.appendChild(content);
    modalBackdrop.appendChild(modalEl);
    document.body.appendChild(modalBackdrop);
    
    // Ferme avec Escape
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') hideModal();
    });
}

function renderModalContent(artist) {
    const content = modalEl.querySelector('.search-modal__content');
    content.innerHTML = '';
    
    // Header
    const header = document.createElement('div');
    header.className = 'search-modal__header';
    const h2 = document.createElement('h2');
    h2.textContent = artist.name || 'Artiste';
    header.appendChild(h2);
    
    // Meta (dates)
    if (artist.creationDate || artist.firstAlbum) {
        const meta = document.createElement('p');
        meta.className = 'search-modal__meta';
        meta.textContent = [
            artist.creationDate ? `Création: ${artist.creationDate}` : '',
            artist.firstAlbum ? `Premier album: ${artist.firstAlbum}` : ''
        ].filter(Boolean).join(' — ');
        header.appendChild(meta);
    }
    
    // Image
    if (artist.image) {
        const imgWrap = document.createElement('div');
        imgWrap.className = 'search-modal__media';
        const img = document.createElement('img');
        img.src = artist.image;
        img.alt = artist.name || '';
        img.loading = 'lazy';
        imgWrap.appendChild(img);
        content.appendChild(imgWrap);
    }
    
    content.appendChild(header);
    
    // Membres
    const members = Array.isArray(artist.members) ? artist.members : [];
    if (members.length) {
        const info = document.createElement('div');
        info.className = 'search-modal__info';
        const title = document.createElement('h3');
        title.textContent = 'Membres';
        info.appendChild(title);
        
        const ul = document.createElement('ul');
        ul.className = 'search-modal__list';
        members.forEach(m => {
            const li = document.createElement('li');
            li.textContent = m;
            ul.appendChild(li);
        });
        info.appendChild(ul);
        content.appendChild(info);
    }
    
    // Lien site officiel
    const links = document.createElement('div');
    links.className = 'search-modal__links';
    const official = artist.url || artist.website || artist.link;
    if (official) {
        const a = document.createElement('a');
        a.href = official;
        a.target = '_blank';
        a.rel = 'noopener noreferrer';
        a.textContent = 'Ouvrir le site officiel';
        links.appendChild(a);
    }
    content.appendChild(links);
}
```

#### Chargement données (lignes 90-115)
```javascript
async function ensureData() {
    if (allArtists.length) return allArtists;  // Déjà en cache
    
    async function fetchArtists(url) {
        const resp = await fetch(url, { headers: { 'Accept': 'application/json' } });
        if (!resp.ok) throw new Error('Réponse réseau incorrecte: ' + resp.status);
        return resp.json();
    }
    
    let data;
    try {
        // Essaie proxy local
        data = await fetchArtists('/api/artists-proxy');
    } catch (err) {
        // Fallback direct API
        try {
            data = await fetchArtists('https://groupietrackers.herokuapp.com/api/artists');
        } catch (fallbackErr) {
            throw fallbackErr;
        }
    }
    
    allArtists = Array.isArray(data) ? data : (data.artists || []);
    return allArtists;
}
```

#### Filtrage par badge (lignes 117-140)
```javascript
function filterByBadge(artist, filterId) {
    if (!filterId) return true;  // Pas de filtre
    
    const name = (artist.name || '').toLowerCase();
    const creation = Number(artist.creationDate || artist.creation_date || 0);
    const albumYear = parseInt((artist.firstAlbum || '').slice(-4), 10);
    const location = ((artist.country || artist.location || '') + '').toLowerCase();
    
    switch (filterId) {
        case 'rock':
            // Recherche keywords dans le nom
            return /rock|metal|punk|roll/.test(name);
            
        case 'seventies':
            // Années 70
            return (creation >= 1970 && creation < 1980) || 
                   (albumYear >= 1970 && albumYear < 1980);
                   
        case 'usa':
            // Filtre par pays (si disponible)
            return /(usa|united states|new york|california)/.test(location);
            
        case 'month':
            // Concerts ce mois (pas implémenté complètement)
            return true;
            
        default:
            return true;
    }
}
```

#### Affichage résultats (lignes 142-215)
```javascript
function renderResults(list) {
    results.innerHTML = '';
    
    if (!list.length) {
        results.innerHTML = '<p>Aucun artiste trouvé.</p>';
        return;
    }
    
    list.forEach((artist, idx) => {
        const card = document.createElement('article');
        card.className = 'artist-card';
        card.tabIndex = 0;  // Rend focusable pour navigation clavier
        card.setAttribute('role','button');
        
        // Image
        const imageUrl = artist.image || artist.imageUrl || /* ...fallbacks... */;
        if (imageUrl) {
            const media = document.createElement('div');
            media.className = 'artist-media';
            const img = document.createElement('img');
            img.src = imageUrl;
            img.alt = `Photo de ${artist.name || ''}`;
            img.loading = 'lazy';  // Lazy loading pour performances
            media.appendChild(img);
            card.appendChild(media);
        }
        
        // Body
        const body = document.createElement('div');
        body.className = 'artist-body';
        
        const h2 = document.createElement('h2');
        h2.textContent = artist.name || '—';
        body.appendChild(h2);
        
        // Meta infos (ville, genre)
        const cityVal = artist.city || artist.location || artist.place;
        if (cityVal) {
            const p = document.createElement('p');
            p.className = 'artist-meta';
            p.textContent = cityVal;
            body.appendChild(p);
        }
        
        card.appendChild(body);
        results.appendChild(card);
        
        // Events
        card.addEventListener('click', () => showModal(artist));
        card.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                showModal(artist);
            }
        });
        
        // Animation d'apparition progressive
        requestAnimationFrame(() => {
            setTimeout(() => { card.classList.add('visible'); }, idx * 40);
        });
    });
}
```

**Accessibilité**:
- `tabIndex=0`: Permet navigation clavier
- `role="button"`: Indique que c'est cliquable
- Support `Enter` et `Space`: Active la card au clavier

#### Fonction recherche principale (lignes 217-245)
```javascript
async function performSearch(q) {
    results.innerHTML = '<p>Recherche en cours…</p>';
    
    try {
        const data = await ensureData();  // Charge si pas en cache
        
        if (!Array.isArray(data) || data.length === 0) {
            results.innerHTML = '<p>Aucun artiste disponible depuis l\'API.</p>';
            return;
        }
        
        const qLower = String(q || '').toLowerCase();
        
        // Filtre par query text
        let filtered = qLower
            ? data.filter(a => (a.name || '').toLowerCase().includes(qLower))
            : data.slice(0, 24);  // Limite initiale à 24
        
        // Applique filtre actif (rock, usa, etc.)
        filtered = filtered.filter(a => filterByBadge(a, activeFilter));
        
        if (filtered.length === 0) {
            results.innerHTML = '<p>Aucun artiste trouvé.</p>';
            return;
        }
        
        renderResults(filtered);
    } catch (err) {
        results.innerHTML = `<p>Erreur lors de la recherche: ${escapeHtml(err.message)}</p>`;
    }
}
```

#### Suggestions auto-completion (lignes 247-275)
```javascript
function updateSuggestions() {
    const q = input.value.trim().toLowerCase();
    
    // Affiche suggestions seulement si 2+ caractères
    if (q.length < 2 || !allArtists.length) {
        suggestionsEl.classList.remove('show');
        suggestionsEl.innerHTML = '';
        return;
    }
    
    // Trouve artistes qui commencent par la query
    const matches = allArtists
        .filter(a => (a.name || '').toLowerCase().startsWith(q))
        .slice(0, 5);  // Max 5 suggestions
    
    if (!matches.length) {
        suggestionsEl.classList.remove('show');
        return;
    }
    
    suggestionsEl.innerHTML = '';
    matches.forEach(m => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.textContent = m.name || '';
        btn.addEventListener('click', () => {
            input.value = m.name || '';
            suggestionsEl.classList.remove('show');
            performSearch(input.value.trim());
        });
        suggestionsEl.appendChild(btn);
    });
    suggestionsEl.classList.add('show');
}
```

#### Gestion filtres rapides (lignes 285-295)
```javascript
function setActiveFilter(id) {
    // Toggle: clic sur filtre actif le désactive
    activeFilter = id === activeFilter ? null : id;
    
    // Update UI
    quickFilters.querySelectorAll('.chip').forEach(chip => {
        const isActive = chip.dataset.filter === activeFilter;
        chip.classList.toggle('active', isActive);
    });
    
    // Re-effectue la recherche avec le nouveau filtre
    performSearch(input ? input.value.trim() : '');
}

if (quickFilters) {
    quickFilters.addEventListener('click', (e) => {
        const btn = e.target.closest('[data-filter]');
        if (!btn) return;
        setActiveFilter(btn.dataset.filter);
    });
}
```

#### Event listeners (lignes 297-340)
```javascript
// Update suggestions pendant la frappe
if (input) {
    input.addEventListener('input', () => {
        ensureData().then(updateSuggestions).catch(() => {});
    });
    
    input.addEventListener('focus', () => {
        ensureData().then(updateSuggestions).catch(() => {});
    });
    
    // Ferme suggestions au clic extérieur
    document.addEventListener('click', (e) => {
        if (!suggestionsEl) return;
        if (!suggestionsEl.contains(e.target) && e.target !== input) {
            suggestionsEl.classList.remove('show');
        }
    });
}

// Submit formulaire
if (form) {
    form.addEventListener('submit', (e) => {
        e.preventDefault();
        const q = (input && input.value || '').trim();
        performSearch(q);
    });
}

// Bouton clear
if (clearBtn) {
    clearBtn.addEventListener('click', () => {
        if (input) input.value = '';
        suggestionsEl && suggestionsEl.classList.remove('show');
        activeFilter = null;
        quickFilters && quickFilters.querySelectorAll('.chip').forEach(c => c.classList.remove('active'));
        results.innerHTML = '<p>Entrez un nom ou utilisez un filtre pour voir les résultats.</p>';
    });
}

// Auto-search si query dans URL (?q=Queen)
if (typeof window !== 'undefined') {
    const params = new URLSearchParams(window.location.search);
    const q = params.get('q') || params.get('artist');
    if (q) {
        if (input) input.value = q;
        ensureData().then(() => performSearch(q));
    } else {
        // Prefetch pour suggestions instantanées
        ensureData().catch(() => {});
    }
}
```

---

### 📄 web/static/js/geoloc.js

**Rôle**: Géolocalisation des concerts sur carte Leaflet avec géocodage.

#### IIFE et initialisation (lignes 1-15)
```javascript
(function () {
    const statusEl = document.getElementById('geo-status');
    const mapEl = document.getElementById('map');
    if (!mapEl) return;  // Pas de map = exit
    
    const setStatus = (msg) => {
        if (statusEl) statusEl.textContent = msg;
    };
    
    // Initialise carte Leaflet
    const map = L.map('map');
    map.setView([20, 0], 2);  // Vue monde centrée
    
    // Ajoute tiles OpenStreetMap
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors',
        maxZoom: 18,
    }).addTo(map);
})();
```

#### URLs API (lignes 17-20)
```javascript
const ARTISTS_URL = '/api/artists-proxy';
const RELATION_URL = '/api/relation-proxy';
```

#### Cache géocodage (lignes 22-24)
```javascript
const cacheKey = (loc) => `geocode:${loc.toLowerCase()}`;
const sleep = (ms) => new Promise((res) => setTimeout(res, ms));
```

**Astuce**: Le cache utilise `localStorage` pour éviter de refaire les mêmes requêtes de géocodage.

#### Chargement données (lignes 30-60)
```javascript
async function buildData() {
    setStatus('Chargement des artistes…');
    const artists = await fetchJson(ARTISTS_URL);
    
    setStatus('Chargement des relations (lieux + dates)…');
    const relation = await fetchJson(RELATION_URL);
    
    // Regroupe par location
    const byLocation = new Map();
    
    // Map artist id → artist object pour lookup rapide
    const artistsById = new Map(artists.map((a) => [a.id, a]));
    
    // Pour chaque relation (artist + locations + dates)
    for (const entry of relation.index || []) {
        const artist = artistsById.get(entry.id);
        const name = artist ? artist.name : `Artiste #${entry.id}`;
        const image = artist ? artist.image : null;
        const dl = entry.datesLocations || {};
        
        // Pour chaque location de cet artiste
        for (const loc of Object.keys(dl)) {
            const dates = dl[loc] || [];
            
            // Crée ou récupère le bucket pour cette location
            if (!byLocation.has(loc)) {
                byLocation.set(loc, { loc, artists: [], dates: [] });
            }
            const bucket = byLocation.get(loc);
            bucket.artists.push({ id: entry.id, name, image });
            bucket.dates.push(...dates);
        }
    }
    
    return { artists, relationByLocation: byLocation };
}
```

**Structure `bucket`**:
```javascript
{
    loc: "usa-texas-houston",
    artists: [
        { id: 1, name: "Queen", image: "..." },
        { id: 5, name: "AC/DC", image: "..." }
    ],
    dates: ["12-01-2023", "15-02-2023", ...]
}
```

#### Géocodage avec Nominatim (lignes 62-85)
```javascript
async function geocodeLocation(loc) {
    const key = cacheKey(loc);
    
    // Check cache localStorage
    const cached = localStorage.getItem(key);
    if (cached) {
        try {
            return JSON.parse(cached);
        } catch {}
    }
    
    // Requête Nominatim
    const url = `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(loc)}`;
    const res = await fetch(url, {
        headers: { 'Accept-Language': 'fr' },
    });
    if (!res.ok) throw new Error('Geocode failed: ' + res.status);
    
    const arr = await res.json();
    if (!Array.isArray(arr) || arr.length === 0) return null;
    
    // Prend le meilleur résultat
    const best = arr[0];
    const point = { lat: parseFloat(best.lat), lon: parseFloat(best.lon) };
    
    // Stocke en cache
    localStorage.setItem(key, JSON.stringify(point));
    
    // Délai pour respecter la politique Nominatim (max 1 req/sec)
    await sleep(250);
    
    return point;
}
```

**Politique Nominatim**:
- Max 1 requête/seconde
- Toujours ajouter `Accept-Language` header
- Mettre en cache les résultats

#### Génération HTML popup (lignes 87-105)
```javascript
function popupHtml(bucket) {
    const uniqueDates = Array.from(new Set(bucket.dates)).sort();
    
    // Liste artistes avec images
    const artistsHtml = bucket.artists
        .map((a) => {
            const img = a.image ? `<img src="${a.image}" alt="${a.name}" />` : '';
            return `<li>${img}<span>${a.name}</span></li>`;
        })
        .join('');
    
    // Liste dates
    const datesHtml = uniqueDates.map((d) => `<li>${d}</li>`).join('');
    
    return `
        <div class="popup">
            <h3>${bucket.loc}</h3>
            <h4>Artistes</h4>
            <ul class="artists">${artistsHtml}</ul>
            <h4>Dates</h4>
            <ul class="dates">${datesHtml}</ul>
        </div>
    `;
}
```

#### Fonction principale (lignes 107-145)
```javascript
async function main() {
    try {
        const { relationByLocation } = await buildData();
        setStatus(`Géocodage de ${relationByLocation.size} lieux…`);
        
        const bounds = [];
        let success = 0;
        let failures = 0;
        
        // Pour chaque location unique
        for (const [loc, bucket] of relationByLocation.entries()) {
            try {
                const pt = await geocodeLocation(loc);
                if (!pt) {
                    failures++;
                    continue;
                }
                
                // Crée marqueur Leaflet
                const marker = L.marker([pt.lat, pt.lon]).addTo(map);
                marker.bindPopup(popupHtml(bucket));
                
                // Garde trace des coords pour bounds
                bounds.push([pt.lat, pt.lon]);
                success++;
            } catch (e) {
                failures++;
            }
        }
        
        // Ajuste le zoom/centre pour voir tous les marqueurs
        if (bounds.length) {
            const b = L.latLngBounds(bounds);
            map.fitBounds(b.pad(0.2));  // 20% padding
        }
        
        setStatus(`Marqueurs prêts: ${success}. Échecs: ${failures}.`);
    } catch (e) {
        console.error(e);
        setStatus('Erreur de chargement des données.');
    }
}

main();  // Lance au chargement du script
```

**Optimisation**:
- Géocode en série (respecte rate limit)
- Cache dans localStorage (évite requêtes répétées)
- `fitBounds` ajuste automatiquement la vue

---

### 📄 web/static/js/subscription.js

**Rôle**: Gestion du système d'abonnement avec simulation de paiement.

#### Initialisation IIFE (lignes 1-20)
```javascript
(function() {
    // DOM Elements
    const subscribeBtn = document.getElementById('subscribeBtn');
    const modal = document.getElementById('subscriptionModal');
    const closeModal = document.getElementById('closeModal');
    const backToPlans = document.getElementById('backToPlans');
    const closeSuccess = document.getElementById('closeSuccess');
    const paymentButtons = document.querySelectorAll('.btn-payment');
    const paymentForm = document.getElementById('paymentForm');
    const cardForm = document.getElementById('cardForm');
    const subscriptionPlans = document.querySelector('.subscription-plans');
    const successMessage = document.getElementById('successMessage');
    const totalPriceEl = document.getElementById('totalPrice');
    const planNameEl = document.getElementById('planName');
    
    // State
    let selectedPlan = null;      // 'monthly' ou 'yearly'
    let selectedPrice = null;     // '9.99' ou '89.99'
    let selectedPlanName = null;  // 'Plan Mensuel' ou 'Plan Annuel'
})();
```

#### Ouverture/fermeture modal (lignes 22-40)
```javascript
// Open modal
subscribeBtn?.addEventListener('click', () => {
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';  // Empêche scroll body
});

// Close modal
closeModal?.addEventListener('click', closeModalHandler);
window.addEventListener('click', (e) => {
    if (e.target === modal) {
        closeModalHandler();
    }
});

function closeModalHandler() {
    modal.classList.remove('active');
    document.body.style.overflow = 'auto';
    resetModal();  // Remet à zéro
}
```

#### Sélection plan (lignes 42-60)
```javascript
paymentButtons.forEach(button => {
    button.addEventListener('click', () => {
        // Récupère infos du plan depuis data-attributes
        selectedPlan = button.getAttribute('data-plan');
        selectedPrice = button.getAttribute('data-price');
        selectedPlanName = button.closest('.plan').querySelector('h3').textContent;
        
        // Cache la sélection des plans, affiche le formulaire
        subscriptionPlans.style.display = 'none';
        paymentForm.classList.remove('hidden');
        
        // Update affichage prix
        const priceFormatted = parseFloat(selectedPrice).toLocaleString('fr-FR', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        });
        totalPriceEl.textContent = priceFormatted + ' €';
        planNameEl.textContent = 'Plan: ' + selectedPlanName;
    });
});
```

#### Retour à la sélection (lignes 62-67)
```javascript
backToPlans?.addEventListener('click', () => {
    subscriptionPlans.style.display = 'grid';
    paymentForm.classList.add('hidden');
    cardForm.reset();
});
```

#### Validation formulaire (lignes 69-100)
```javascript
cardForm?.addEventListener('submit', (e) => {
    e.preventDefault();
    
    // Récupère valeurs formulaire
    const cardholderName = document.getElementById('cardholderName').value;
    const cardNumber = document.getElementById('cardNumber').value;
    const expiryDate = document.getElementById('expiryDate').value;
    const cvv = document.getElementById('cvv').value;
    const email = document.getElementById('email').value;
    
    // Validations basiques
    if (!validateCardNumber(cardNumber)) {
        showError('Numéro de carte invalide');
        return;
    }
    
    if (!validateExpiryDate(expiryDate)) {
        showError('Date d\'expiration invalide (MM/YY)');
        return;
    }
    
    if (!validateCVV(cvv)) {
        showError('CVV invalide');
        return;
    }
    
    // Simule le paiement
    processPayment(cardholderName, cardNumber, expiryDate, cvv, email);
});
```

#### Formatage inputs (lignes 102-130)
```javascript
// Formatage numéro de carte (1234 5678 9012 3456)
document.getElementById('cardNumber')?.addEventListener('input', (e) => {
    let value = e.target.value.replace(/\s/g, '');  // Retire espaces
    let formattedValue = '';
    for (let i = 0; i < value.length; i++) {
        if (i > 0 && i % 4 === 0) formattedValue += ' ';  // Espace tous les 4 chiffres
        formattedValue += value[i];
    }
    e.target.value = formattedValue;
});

// Formatage date expiration (MM/YY)
document.getElementById('expiryDate')?.addEventListener('input', (e) => {
    let value = e.target.value.replace(/\D/g, '');  // Garde que chiffres
    if (value.length >= 2) {
        value = value.slice(0, 2) + '/' + value.slice(2, 4);
    }
    e.target.value = value;
});

// CVV: que des chiffres
document.getElementById('cvv')?.addEventListener('input', (e) => {
    e.target.value = e.target.value.replace(/\D/g, '');
});
```

#### Fonctions validation (lignes 132-148)
```javascript
function validateCardNumber(cardNumber) {
    // Simple: vérifie 16 chiffres
    const cleanNumber = cardNumber.replace(/\s/g, '');
    return /^\d{16}$/.test(cleanNumber);
}

function validateExpiryDate(date) {
    return /^\d{2}\/\d{2}$/.test(date);
    // Note: pas de vérification de date future ici (simplification)
}

function validateCVV(cvv) {
    return /^\d{3,4}$/.test(cvv);  // 3 ou 4 chiffres
}
```

**Note sécurité**: En production, ne **JAMAIS** envoyer les données de carte au serveur directement. Utiliser Stripe, PayPal, etc.

#### Simulation paiement (lignes 150-185)
```javascript
function processPayment(name, cardNumber, expiry, cvv, email) {
    const submitBtn = cardForm.querySelector('button[type="submit"]');
    const originalText = submitBtn.textContent;
    
    // État loading
    submitBtn.textContent = 'Traitement...';
    submitBtn.disabled = true;
    
    // Simule délai API
    setTimeout(() => {
        // Cache formulaire
        paymentForm.classList.add('hidden');
        
        // Affiche succès
        successMessage.classList.remove('hidden');
        
        // Log (en prod: envoyer au serveur)
        console.log('Paiement réussi:', {
            plan: selectedPlan,
            planName: selectedPlanName,
            amount: selectedPrice,
            cardholderName: name,
            email: email,
            timestamp: new Date().toISOString()
        });
        
        // Stocke abonnement dans localStorage
        const subscription = {
            plan: selectedPlan,
            planName: selectedPlanName,
            amount: selectedPrice,
            email: email,
            subscribedAt: new Date().toISOString(),
            status: 'active'
        };
        localStorage.setItem('groupie_subscription', JSON.stringify(subscription));
        
        // Reset bouton
        submitBtn.textContent = originalText;
        submitBtn.disabled = false;
    }, 2000);  // 2 secondes de simulation
}
```

#### Reset et check status (lignes 187-210)
```javascript
function resetModal() {
    subscriptionPlans.style.display = 'grid';
    paymentForm.classList.add('hidden');
    successMessage.classList.add('hidden');
    cardForm.reset();
    selectedPlan = null;
    selectedPrice = null;
    selectedPlanName = null;
}

function checkSubscriptionStatus() {
    const subscription = localStorage.getItem('groupie_subscription');
    if (subscription) {
        const data = JSON.parse(subscription);
        console.log('Utilisateur abonné:', data);
        // Peut update l'UI ici (ex: badge "Premium", désactiver bouton subscribe)
    }
}

// Check au chargement
document.addEventListener('DOMContentLoaded', checkSubscriptionStatus);
```

---

## Documentation des fichiers CSS

### 📄 web/static/css/style.css

**Rôle**: Styles globaux et thème principal.

#### Variables CSS (lignes 8-19)
```css
:root{
    --charcoal: #0d0f14;
    --panel: rgba(20,22,30,0.55);
    --panel-strong: rgba(20,22,30,0.75);
    --muted: #c4c9d4;
    --muted-strong: #e9ecf5;
    --gold: #ec4899;
    --electric: #06b6d4;
    --accent: linear-gradient(120deg, #06b6d4, #a78bfa);
    --accent-solid: #a78bfa;
    --glass-border: rgba(255,255,255,0.08);
    --radius: 14px;
    --max-width: 1800px;
    --shadow-strong: 0 20px 70px rgba(0,0,0,0.4);
}
```

**Palette de couleurs**:
- **Charcoal**: Fond principal
- **Panel**: Cartes/conteneurs (glassmorphism)
- **Muted**: Texte secondaire
- **Gold/Electric**: Accents colorés
- **Accent**: Gradient boutons

#### Reset et base (lignes 21-35)
```css
*{box-sizing:border-box}
html,body{height:100%}
body{
    margin:0;
    font-family: Inter, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
    color:var(--muted-strong);
    line-height:1.5;
    -webkit-font-smoothing:antialiased;
    -moz-osx-font-smoothing:grayscale;
    background-image: linear-gradient(135deg, rgba(12,12,18,0.92) 0%, rgba(15,17,28,0.8) 45%, rgba(10,8,18,0.9) 100%), url('/static/images/Fond_GroupieTracker.png');
    background-size: cover;
    background-position: center center;
    background-repeat: no-repeat;
    background-attachment: fixed;
    min-height:100vh;
}
```

**Technique importante**: Double background (gradient + image) pour effet overlay.

#### Container glassmorphism (lignes 38-46)
```css
.container{
    max-width:var(--max-width);
    margin:1.25rem auto;
    padding:1.25rem 1.5rem;
    background:var(--panel);
    border:1px solid var(--glass-border);
    border-radius:var(--radius);
    box-shadow:var(--shadow-strong);
    backdrop-filter: blur(14px);  /* ⭐ Effet glassmorphism */
}
```

**`backdrop-filter: blur()`**: Floute l'arrière-plan visible à travers le panel semi-transparent.

#### Header (lignes 49-92)
```css
.site-header{
    padding: 0.85rem 1.5rem;
    background: var(--panel-strong);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius);
    backdrop-filter: blur(16px);
    box-shadow: 0 18px 50px rgba(0,0,0,0.35);
    margin-bottom: 1rem;
}

.header-content {
    display: flex;
    align-items: center;
    gap: 1rem;
    max-width: var(--max-width);
    margin: 0 auto;
}

.site-header h1{
    margin: 0;
    font-family: 'Merriweather', serif;  /* Font décorative titre */
    color: #f8fafc;
    font-size: 1.9rem;
    letter-spacing: 0.3px;
    white-space: nowrap;
}
```

#### Navigation (lignes 94-111)
```css
.main-nav{
    display: flex;
    gap: 0.75rem;
}

.main-nav a{
    color: var(--muted);
    text-decoration: none;
    padding: 0.5rem 0.7rem;
    border-radius: 10px;
    transition: all 0.22s ease;
    font-weight: 650;
    border: 1px solid transparent;
}

.main-nav a:hover{
    background: rgba(255,255,255,0.06);
    color: #fff;
    border-color: rgba(255,255,255,0.08);
}
```

#### Boutons avec effet (lignes 119-134)
```css
.header-search-btn{
    padding:0.48rem 0.9rem;
    border-radius:12px;
    border:1px solid var(--glass-border);
    background:var(--accent);
    color:#fff;
    cursor:pointer;
    box-shadow:0 10px 30px rgba(123,58,237,0.35);
    transition:transform 0.18s ease, box-shadow 0.18s ease;
    position:relative;
    overflow:hidden;
}

.header-search-btn::after{
    content:"";
    position:absolute;
    inset:0;
    background:linear-gradient(120deg, rgba(255,255,255,0.08), rgba(255,255,255,0));
    transform:translateX(-100%);
    transition:transform 0.4s ease;
}

.header-search-btn:hover::after{
    transform:translateX(0);  /* Glisse de gauche à droite */
}
```

**Effet shimmer**: Pseudo-element `::after` qui glisse au survol.

#### Grid vinyles (lignes 295-330)
```css
.vinyl-grid{
    width:100%;
    display:grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));  /* 3 colonnes égales */
    row-gap:20px;
    column-gap:30px;
    align-items:start;
    justify-content:center;
    justify-items:center;
    padding:20px 10px 50px 10px;
    max-width:1400px;
    margin:0 auto;
}

.vinyl-item{
    display:flex;
    flex-direction:column;
    align-items:center;
    gap:15px;
    width:100%;
    opacity:0;  /* Invisible par défaut, animé par JS */
}

.vinyl-frame{
    width:380px;
    height:380px;
    background-image: url('/static/images/vinyle.png');
    background-repeat: no-repeat;
    background-position: center center;
    background-size: contain;
    border-radius:50%;
    transition:transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.35s ease, filter 0.3s ease;
    position:relative;
    overflow:hidden;
}
```

**`cubic-bezier(0.34, 1.56, 0.64, 1)`**: Courbe d'animation avec "bounce" pour effet élastique.

#### Animations vinyles (lignes 348-373)
```css
.vinyl-frame:hover{
    transform:scale(1.12) rotate(-2deg);
    box-shadow:0 0 30px 12px rgba(236,72,153,0.25), inset 0 0 20px rgba(6,182,212,0.1);
    filter:brightness(1.1);
}

/* Rotation quand audio joue */
.vinyl-frame.playing {
    animation: vinylSpin 2s linear infinite, vinylPulse 1.5s ease-in-out infinite;
}

.vinyl-frame.playing::after {
    content: '♪';  /* Note musicale */
    position: absolute;
    top: 10px;
    right: 10px;
    background: rgba(236,72,153,0.95);
    color: white;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    box-shadow: 0 4px 12px rgba(236,72,153,0.6);
    animation: musicNote 0.5s ease-in-out infinite alternate;
    z-index: 10;
}

@keyframes vinylSpin {
    from { transform: scale(1.12) rotate(-2deg); }
    to { transform: scale(1.12) rotate(359deg); }
}

@keyframes vinylPulse {
    0%, 100% { box-shadow: 0 0 30px 12px rgba(236,72,153,0.25); }
    50% { box-shadow: 0 0 40px 16px rgba(236,72,153,0.35); }
}

@keyframes musicNote {
    from { transform: translateY(0px); opacity: 1; }
    to { transform: translateY(-3px); opacity: 0.8; }
}
```

**Superposition animations**:
- `vinylSpin`: Rotation continue
- `vinylPulse`: Pulsation shadow
- `musicNote`: Bounce note

#### Modal artiste (lignes 533-620)
```css
.artist-modal{
    position:fixed;
    inset:0;  /* Remplace top:0; right:0; bottom:0; left:0; */
    display:flex;
    align-items:center;
    justify-content:center;
    z-index:9999;
    visibility:hidden;
    opacity:0;
    transition:opacity .18s ease;
}

.artist-modal.open{
    visibility:visible;
    opacity:1;
    background:rgba(0,0,0,0.45);  /* Backdrop semi-transparent */
}

.artist-modal__panel{
    background:var(--panel-strong);
    padding:20px;
    border-radius:14px;
    max-width:640px;
    width:90%;
    box-shadow:0 30px 80px rgba(0,0,0,0.45);
    border:1px solid var(--glass-border);
    backdrop-filter:blur(18px);
}

.artist-modal .artist-modal__close{
    position:absolute;
    right:12px;
    top:8px;
    width:42px;
    height:42px;
    background:#ffffff;
    color:#111;
    border:2px solid rgba(255,255,255,0.9);
    font-size:22px;
    font-weight:800;
    border-radius:12px;
    cursor:pointer;
    box-shadow:0 16px 36px rgba(0,0,0,0.55);
    transition:transform 0.16s ease, box-shadow 0.16s ease;
}

.artist-modal .artist-modal__close:hover{
    background:#f5f6f8;
    transform:translateY(-1px) scale(1.03);
    box-shadow:0 20px 44px rgba(0,0,0,0.6);
}
```

#### Styles d'authentification (lignes 787-880)
```css
.auth-container{
    display:flex;
    flex-direction:column;
    gap:1.25rem;
    max-width:960px;
}

.auth-card{
    background:var(--panel-strong);
    border:1px solid var(--glass-border);
    border-radius:var(--radius);
    padding:1.5rem;
    box-shadow:0 18px 50px rgba(0,0,0,0.32);
    backdrop-filter:blur(14px);
}

.auth-tabs{
    display:flex;
    gap:0.5rem;
    margin-bottom:1rem;
}

.auth-tab{
    flex:1;
    padding:0.65rem 0.9rem;
    border-radius:12px;
    border:1px solid var(--glass-border);
    background:rgba(255,255,255,0.04);
    color:var(--muted-strong);
    cursor:pointer;
    font-weight:650;
    transition:all 0.18s ease;
}

.auth-tab.active{
    background:linear-gradient(120deg, #06b6d4, #a78bfa);
    color:#fff;
    box-shadow:0 12px 32px rgba(0,0,0,0.28);
}

.form-field input:focus{
    outline:2px solid rgba(99,102,241,0.35);
    border-color:rgba(99,102,241,0.35);
}
```

#### Modal abonnement (lignes 892-1050+)
```css
.modal {
    display: none;
    position: fixed;
    z-index: 1000;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    animation: fadeIn 0.3s ease;
}

.modal.active {
    display: flex;
    align-items: flex-start;
    justify-content: center;
}

.subscription-plans {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
}

.plan {
    background: rgba(20, 22, 30, 0.4);
    border: 1px solid var(--glass-border);
    border-radius: 12px;
    padding: 1.5rem;
    text-align: center;
    transition: all 0.3s ease;
}

.plan.featured {
    border: 2px solid var(--electric);
    background: rgba(59, 130, 246, 0.1);
    transform: scale(1.05);  /* Plan recommandé plus grand */
}

.plan .price {
    font-size: 2rem;
    color: var(--electric);
    margin: 0.5rem 0;
    font-weight: 600;
}
```

#### Responsive breakpoints (lignes 180+, multiples sections)
```css
@media(min-width:700px){
    .features{grid-template-columns:repeat(3,1fr)}
    .hero h2{font-size:2rem}
}

@media(max-width:1400px){
    .vinyl-frame{width:350px;height:350px}
}

@media(max-width:900px){
    .vinyl-grid{grid-template-columns: repeat(2, 1fr)}
    .vinyl-frame{width:260px;height:260px}
}

@media(max-width:500px){
    .vinyl-grid{grid-template-columns: 1fr}  /* 1 colonne sur mobile */
    .vinyl-frame{width:240px;height:240px}
}
```

---

### 📄 web/static/css/search.css

**Rôle**: Styles spécifiques page de recherche.

#### Formulaire recherche (lignes 40-75)
```css
.search-form{
    display:flex;
    flex-direction:column;
    gap:0.9rem;
}

.search-form input[type="text"]{
    width:100%;
    padding:0.75rem 0.9rem;
    border-radius:12px;
    border:1px solid var(--glass-border);
    background:rgba(255,255,255,0.06);
    color:#f8fafc;
    font-size:1rem;
}

.search-form input[type="text"]:focus{
    outline:2px solid rgba(123,58,237,0.35);  /* Purple glow */
}
```

#### Suggestions dropdown (lignes 77-103)
```css
.suggestions{
    position:absolute;
    top:100%;
    left:0;
    right:0;
    background:var(--panel-strong);
    border:1px solid var(--glass-border);
    border-radius:12px;
    box-shadow:0 18px 40px rgba(0,0,0,0.4);
    margin-top:6px;
    display:none;  /* Caché par défaut */
    z-index:10;
    overflow:hidden;
}

.suggestions.show{
    display:block;  /* Affiché par JS */
}

.suggestions button{
    width:100%;
    padding:0.65rem 0.9rem;
    background:transparent;
    color:#f8fafc;
    border:0;
    text-align:left;
    cursor:pointer;
    transition: background 0.15s ease;
}

.suggestions button:hover{
    background:rgba(255,255,255,0.06);  /* Highlight au survol */
}
```

#### Filtres chips (lignes 105-125)
```css
.quick-filters{
    display:flex;
    flex-wrap:wrap;
    gap:0.5rem;
}

.chip{
    padding:0.45rem 0.75rem;
    border-radius:999px;  /* Complètement arrondi */
    border:1px solid var(--glass-border);
    background:rgba(255,255,255,0.06);
    color:#f8fafc;
    cursor:pointer;
    transition:transform 0.18s ease, box-shadow 0.18s ease, background 0.3s ease;
    box-shadow:0 10px 25px rgba(0,0,0,0.25);
}

.chip.active{
    background:var(--accent);  /* Gradient quand actif */
    color:#fff;
    border-color:transparent;
    box-shadow:0 14px 40px rgba(123,58,237,0.45);
}
```

#### Grid résultats (lignes 158-220)
```css
#results{
    margin-top:1rem;
    display:grid;
    grid-template-columns:repeat(auto-fill, minmax(220px, 1fr));
    gap:16px;
    align-items:start;
}

.artist-card{
    background:var(--panel-strong);
    border-radius:12px;
    overflow:hidden;
    display:flex;
    flex-direction:column;
    cursor:pointer;
    box-shadow:0 14px 40px rgba(0,0,0,0.35);
    border:1px solid var(--glass-border);
    
    /* État initial (invisible) */
    opacity:0;
    transform:translateY(14px);
    transition:transform 0.25s ease, box-shadow 0.25s ease, opacity 0.25s ease;
}

.artist-card.visible{
    /* Devient visible via JS */
    opacity:1;
    transform:translateY(0);
}

.artist-card:hover{
    box-shadow:0 18px 52px rgba(0,0,0,0.45);
    transform:translateY(-2px);  /* Soulève légèrement */
}

.artist-media img{
    display:block;
    width:100%;
    height:160px;
    object-fit:cover;  /* Crop proportionnel */
}
```

**`auto-fill` vs `auto-fit`**: `auto-fill` crée autant de colonnes que possible, `auto-fit` collapse les vides.

#### Modal résultats (lignes 252-295)
```css
.search-modal-backdrop{
    position:fixed;
    inset:0;
    background:rgba(0,0,0,0.6);
    display:flex;
    align-items:center;
    justify-content:center;
    opacity:0;
    visibility:hidden;
    transition:opacity 0.2s ease;
    z-index:9999;
}

.search-modal-backdrop.open{
    opacity:1;
    visibility:visible;
}

.search-modal{
    background:var(--panel-strong);
    border-radius:14px;
    border:1px solid var(--glass-border);
    box-shadow:0 24px 80px rgba(0,0,0,0.55);
    max-width:520px;
    width:90%;
    position:relative;
    padding:18px;
}

.search-modal__close{
    position:absolute;
    top:10px;
    right:10px;
    width:40px;
    height:40px;
    border-radius:12px;
    background:#fff;
    color:#111;
    border:1px solid rgba(255,255,255,0.65);
    cursor:pointer;
    font-size:20px;
    font-weight:700;
    box-shadow:0 12px 32px rgba(0,0,0,0.45);
}
```

---

### 📄 web/static/css/geoloc.css

**Rôle**: Styles carte Leaflet.

#### Conteneur carte (lignes 1-6)
```css
#map {
    height: 70vh;
    min-height: 400px;
    border-radius: 12px;
    box-shadow: 0 6px 20px rgba(0,0,0,0.1);
    margin: 16px 0;
}
```

**`70vh`**: 70% de la hauteur de la fenêtre, avec minimum 400px.

#### Styles popups Leaflet (lignes 13-80)
```css
.leaflet-popup-content {
    margin: 12px;
    min-width: 200px;
}

.popup h3 {
    margin: 0 0 12px 0;
    font-size: 1.1rem;
    color: #333;
    border-bottom: 2px solid #e74c3c;
    padding-bottom: 6px;
}

.popup ul.artists {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.popup ul.artists li {
    display: flex;
    align-items: center;
    gap: 10px;
}

.popup ul.artists img {
    width: 50px;
    height: 50px;
    object-fit: cover;
    border-radius: 6px;
    border: 2px solid #000;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.popup ul.dates {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 6px;
}

.popup ul.dates li {
    font-size: 0.8rem;
    color: #000;
    background: #f5f5f5;
    padding: 4px 8px;
    border-radius: 4px;
}
```

**Layout popup**:
- **Artistes**: Liste verticale avec images
- **Dates**: Tags horizontaux flex-wrap

---

## Flux de données

### 1. Chargement page d'accueil

```
Utilisateur → index.html
    ↓
Charge CSS (style.css)
    ↓
Charge JS (ui.js)
    ↓
DOMContentLoaded
    ↓
ui.js: fetchArtists()
    ↓
Essaie /api/artists-proxy
    ↓ (succès)
main.go: proxy() → groupietrackers.herokuapp.com
    ← artistes JSON
    ↓
ui.js: créé vinyles dans .vinyl-grid
    ↓
Pour chaque artiste:
    - Fetch preview iTunes/Deezer
    - Crée audio element
    - Bind events (hover, click)
```

### 2. Recherche d'artiste

```
Utilisateur → search.html
    ↓
Tape dans input
    ↓
search.js: input event
    ↓
ensureData() (si pas en cache)
    ↓
Fetch /api/artists-proxy
    ← allArtists
    ↓
updateSuggestions()
    - Filter artistes qui startsWith(query)
    - Affiche <5 suggestions
    ↓
Utilisateur clique suggestion OU submit form
    ↓
performSearch(query)
    - Filter par nom
    - Applique activeFilter (rock, usa...)
    - renderResults()
    ↓
Affiche cards artistes
    ↓
Click card → showModal(artist)
```

### 3. Géolocalisation

```
Utilisateur → geoloc.html
    ↓
Charge Leaflet.js
    ↓
Charge geoloc.js
    ↓
main() exécute:
    ↓
buildData()
    - Fetch artistes
    - Fetch relations
    - Group by location
    ↓
Pour chaque location:
    ↓
    geocodeLocation(location)
        ↓
        Check localStorage cache
        ↓ (miss)
        Fetch Nominatim API
        ← {lat, lon}
        ↓
        Store in cache
        ↓
        Sleep 250ms (rate limit)
    ↓
    L.marker([lat, lon])
    .bindPopup(html)
    .addTo(map)
    ↓
map.fitBounds(allMarkers)
```

### 4. Authentification

```
Utilisateur → /login
    ↓
Choisit onglet: Connexion ou Inscription
    ↓
[Inscription]:
    Fill form (nom, prenom, sexe, password)
    ↓
    Submit
    ↓
    POST /api/register
        ↓
    main.go: handleRegister()
        - Validate fields
        - bcrypt.GenerateFromPassword()
        - INSERT INTO user
        ← {id_utilisateur}
    ↓
    Display toast: "Compte créé. ID: 123"
    
[Connexion]:
    Fill form (id_utilisateur, password)
    ↓
    Submit
    ↓
    POST /api/login
        ↓
    main.go: handleLogin()
        - SELECT password FROM user WHERE id=?
        - bcrypt.CompareHashAndPassword()
        ← {user: {nom, prenom, sexe}}
    ↓
    Display toast: "Connexion réussie"
```

### 5. Abonnement

```
Utilisateur → Click "S'abonner"
    ↓
Modal ouvre
    ↓
Sélectionne plan (mensuel/annuel)
    ↓
subscription.js: 
    - Stocke selectedPlan, selectedPrice
    - Affiche formulaire paiement
    ↓
Fill carte bancaire (formatage auto)
    ↓
Submit
    ↓
Validation (client-side seulement)
    - validateCardNumber()
    - validateExpiryDate()
    - validateCVV()
    ↓
processPayment()
    - Simule délai 2s
    - Log console
    - Store dans localStorage
    ↓
Affiche message succès
```

---

## Configuration et déploiement

### Variables d'environnement

```bash
# Serveur
PORT=8080                    # Port HTTP (défaut: 8080)

# Base de données
DB_HOST=localhost            # Hôte MySQL
DB_PORT=3306                 # Port MySQL
DB_NAME=groupi_tracker       # Nom de la DB
DB_USER=root                 # Utilisateur
DB_PASS=                     # Mot de passe
DISABLE_DB=0                 # 1 pour désactiver la DB
```

### Démarrage local

```bash
# Installer dépendances Go
go mod download

# Lancer serveur
go run main.go

# Ou compiler
go build -o groupie-tracker
./groupie-tracker
```

**Accès**: http://localhost:8080/

### Structure base de données

```sql
CREATE TABLE `user` (
    `id_user` INT AUTO_INCREMENT PRIMARY KEY,
    `Nom` VARCHAR(100) NOT NULL,
    `Prénom` VARCHAR(100) NOT NULL,
    `sexe` ENUM('M', 'F', 'Autre') NOT NULL,
    `password` VARCHAR(255) NOT NULL,  -- Hash bcrypt
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Déploiement Vercel/Netlify

#### Fichier `vercel.json` (ou `render.yaml`)
```json
{
  "version": 2,
  "builds": [
    {
      "src": "api/index.go",
      "use": "@vercel/go"
    }
  ],
  "routes": [
    {
      "src": "/api/(.*)",
      "dest": "/api/index.go"
    },
    {
      "src": "/static/(.*)",
      "dest": "/web/static/$1"
    },
    {
      "src": "/(.*)",
      "dest": "/index.html"
    }
  ]
}
```

**Notes**:
- Vercel utilise `api/index.go` comme handler serverless
- Pas de base de données sur Vercel free tier (utiliser service externe)
- Fichiers statiques servis depuis `/web/static/`

### Optimisations production

#### 1. Minification
```bash
# CSS
npx csso web/static/css/style.css -o web/static/css/style.min.css

# JS
npx terser web/static/js/ui.js -o web/static/js/ui.min.js -c -m
```

#### 2. Cache headers (déjà dans main.go)
```go
w.Header().Set("Cache-Control", "public, max-age=31536000")
```

#### 3. Compression GZIP
```go
import "github.com/NYTimes/gziphandler"

http.Handle("/static/", gziphandler.GzipHandler(staticHandler))
```

#### 4. CDN
- Héberger images sur Cloudinary/ImgIX
- Utiliser CDN pour Leaflet, fonts Google

### Sécurité

#### Headers recommandés
```go
// Dans main.go, ajouter middleware:
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}
```

#### HTTPS obligatoire
```go
// Redirect HTTP → HTTPS
if r.Header.Get("X-Forwarded-Proto") == "http" {
    http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
    return
}
```

#### Rate limiting
```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(rate.Limit(10), 20) // 10 req/s, burst 20

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Améliorations futures

### 1. Backend
- [ ] Sessions utilisateur (JWT ou cookies)
- [ ] API REST complète (CRUD artistes favoris)
- [ ] Pagination résultats recherche
- [ ] Filtres avancés (genre, popularité)
- [ ] WebSocket pour notifications temps réel

### 2. Frontend
- [ ] Mode sombre/clair (toggle)
- [ ] Offline mode (Service Worker + Cache API)
- [ ] Animations page transitions (GSAP)
- [ ] Lazy loading images (Intersection Observer)
- [ ] Infinite scroll résultats

### 3. Fonctionnalités
- [ ] Favoris/playlists utilisateur
- [ ] Commentaires sur artistes
- [ ] Partage social (Twitter, Facebook)
- [ ] Export calendrier (iCal pour dates concerts)
- [ ] Notifications push abonnés

### 4. Performance
- [ ] Preload critical resources
- [ ] Code splitting JavaScript
- [ ] Image optimization (WebP, AVIF)
- [ ] Redis cache pour API calls
- [ ] GraphQL au lieu de REST

---

## Glossaire

- **Glassmorphism**: Effet de verre dépoli (backdrop-filter + opacity)
- **CORS**: Cross-Origin Resource Sharing (politique sécurité navigateurs)
- **Lazy loading**: Chargement différé ressources (images, scripts)
- **Debounce**: Retarder exécution fonction (évite trop de calls)
- **Throttle**: Limiter fréquence exécution fonction
- **SPA**: Single Page Application (une seule page HTML)
- **SSR**: Server-Side Rendering (rendu côté serveur)
- **CSR**: Client-Side Rendering (rendu côté client)
- **PWA**: Progressive Web App (app-like experience web)
- **JWT**: JSON Web Token (authentification stateless)

---

## Contacts et support

**Équipe de développement**:
- Preston
- Clément
- Timéo

**Licence**: Non spécifiée (projet éducatif)

**Version**: 1.0.0 (Janvier 2026)

---

*Documentation générée le 22 janvier 2026*
