package main

import (
	"auth-service/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

type Claims struct {
	UserID string `json:"user_id"`
	IP     string `json:"ip"`
	jwt.StandardClaims
}

func issueTokens(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	claims := &Claims{
		UserID: userID,
		IP:     r.RemoteAddr,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessToken, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	refreshToken := utils.GenerateRandomBase64String(32)
	refreshTokenHash, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	db, err := utils.InitDB()
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO tokens (user_id, refresh_token, ip_address) VALUES ($1, $2, $3)", userID, string(refreshTokenHash), r.RemoteAddr)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"access_token":"%s","refresh_token":"%s"}`, accessToken, refreshToken)))
}

func refreshTokens(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid", http.StatusBadRequest)
		return
	}

	refreshToken := requestBody["refresh_token"]
	if refreshToken == "" {
		http.Error(w, "required", http.StatusBadRequest)
		return
	}

	db, err := utils.InitDB()
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var storedTokenHash, userID, storedIP string
	err = db.QueryRow("SELECT refresh_token, user_id, ip_address FROM tokens WHERE refresh_token = $1", refreshToken).Scan(&storedTokenHash, &userID, &storedIP)
	if err != nil {
		http.Error(w, "Invalid", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedTokenHash), []byte(refreshToken)); err != nil {
		http.Error(w, "Invalid", http.StatusUnauthorized)
		return
	}

	if storedIP != r.RemoteAddr {
		utils.SendWarningEmail(userID)
	}

	claims := &Claims{
		UserID: userID,
		IP:     r.RemoteAddr,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	newAccessToken, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"access_token":"%s"}`, newAccessToken)))
}

func main() {
	http.HandleFunc("/tokens", issueTokens)
	http.HandleFunc("/refresh", refreshTokens)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
