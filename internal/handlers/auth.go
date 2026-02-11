package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Nom      string `json:"nom"`
	Prenom   string `json:"prenom"`
	Sexe     string `json:"sexe"`
	Password string `json:"password"`
}

var DB *sql.DB

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable"})
		return
	}

	var req user
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	log.Printf("Register attempt: Nom=%q, Prenom=%q, Sexe=%q, PwdLen=%d", req.Nom, req.Prenom, req.Sexe, len(req.Password))

	if req.Nom == "" || req.Prenom == "" || req.Sexe == "" || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing or invalid fields"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	query := "INSERT INTO `user` (`Nom`, `Prénom`, `sexe`, `password`) VALUES (?, ?, ?, ?)"
	result, err := DB.Exec(query, req.Nom, req.Prenom, req.Sexe, string(hash))
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{"message": "user created", "id_utilisateur": id})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database unavailable"})
		return
	}

	var req struct {
		IDUser   int    `json:"id_utilisateur"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.IDUser <= 0 || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing credentials"})
		return
	}

	var storedHash string
	var nom, prenom string
	var sexe string
	row := DB.QueryRow("SELECT `password`, `Nom`, `Prénom`, `sexe` FROM `user` WHERE `id_user` = ?", req.IDUser)
	if err := row.Scan(&storedHash, &nom, &prenom, &sexe); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "login ok",
		"user": map[string]any{
			"id_utilisateur": req.IDUser,
			"nom":            nom,
			"prenom":         prenom,
			"sexe":           sexe,
		},
	})
}
