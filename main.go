package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	database       *database.DB
	jwtSecret      string
	fileserverHits int
}

func main() {
	// by default, godotenv will look for a file named .env in the current directory
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal("Database crashed:", err)
	}
	apiCfg := &apiConfig{
		fileserverHits: 0,
		jwtSecret:      jwtSecret,
		database:       db,
	}
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}

	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/api/healthz", handleReadinessEndpoint)
	mux.HandleFunc("/api/chirps", apiCfg.handleChirps)
	mux.HandleFunc("/api/chirps/{chirpId}", apiCfg.handleIndividualChirp)
	mux.HandleFunc("/admin/metrics", apiCfg.handleMetricsEndpoint)
	mux.HandleFunc("POST /api/users", apiCfg.createUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.generateAccessToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeUser)
	mux.HandleFunc("/api/login", apiCfg.loginUser)
	mux.HandleFunc("/api/reset", apiCfg.handleResetEndpoint)

	log.Println("Starting server on :8080")
	server.ListenAndServe()
}

func (a *apiConfig) revokeUser(w http.ResponseWriter, r *http.Request) {
	authorizationToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	err := a.database.RevokeRefreshToken(string(authorizationToken))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong in the database")
		return
	}

	respondWithJson(w, http.StatusNoContent, nil)
}

func (a *apiConfig) generateAccessToken(w http.ResponseWriter, r *http.Request) {
	type ResponseBody struct {
		Token string `json:"token"`
	}
	authorizationToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	user, err := a.database.GetUserByRefreshToken(string(authorizationToken))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong in the database")
		return
	}
	if len(user.Email) == 0 {
		respondWithError(w, http.StatusUnauthorized, "this user couldn't be authorized")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(time.Hour))),
		Subject:   fmt.Sprint(user.Id),
	})
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := ResponseBody{
		Token: tokenString,
	}
	respondWithJson(w, http.StatusOK, response)
}

func (a *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	type ResponseBody struct {
		Email string `json:"email"`
		Id    int    `json:"id"`
	}
	type RequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}
	token := strings.Split(r.Header.Get("Authorization"), " ")[1]
	claims := jwt.RegisteredClaims{}
	tokenPointer, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(a.jwtSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "the user is not authorized")
		return
	}
	id, err := tokenPointer.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "the token is malformed")
		return
	}
	user, err := a.database.UpdateUser(id, bodyJson.Email, bodyJson.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "we couldn't update the user")
		return
	}

	response := ResponseBody{
		Email: user.Email,
		Id:    user.Id,
	}
	respondWithJson(w, http.StatusOK, response)
}

func (a *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		ExpirationTime int    `json:"expires_in_seconds"`
	}
	type ResponseBody struct {
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		Id           int    `json:"id"`
	}
	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}
	user, err := a.database.LoginUser(bodyJson.Email, bodyJson.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if bodyJson.ExpirationTime == 0 {
		bodyJson.ExpirationTime = 24 * 60 * 60
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(time.Duration(bodyJson.ExpirationTime) * time.Second))),
		Subject:   fmt.Sprint(user.Id),
	})
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := ResponseBody{
		Email:        user.Email,
		Id:           user.Id,
		Token:        tokenString,
		RefreshToken: user.RefreshToken,
	}
	respondWithJson(w, http.StatusOK, response)
}

func (a *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type ResponseBody struct {
		Email string `json:"email"`
		Id    int    `json:"id"`
	}

	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}

	user, err := a.database.CreateUser(bodyJson.Email, bodyJson.Password)
	response := ResponseBody{
		Email: user.Email,
		Id:    user.Id,
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, response)
}

func (a *apiConfig) handleChirps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		a.createChirps(w, r)
	case "GET":
		a.fetchChirps(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *apiConfig) handleIndividualChirp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		a.fetchSingleChirp(w, r)
	// case "DELETE":
	// 	a.fetchChirps(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *apiConfig) fetchSingleChirp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "provide correct id")
		return
	}
	chirp, err := a.database.GetSingleChirp(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	respondWithJson(w, http.StatusOK, chirp)
}

func (a *apiConfig) createChirps(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Body string `json:"body"`
	}

	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(bodyJson.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	token := strings.Split(r.Header.Get("Authorization"), " ")[1]
	claims := jwt.RegisteredClaims{}
	tokenPointer, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(a.jwtSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "the user is not authorized")
		return
	}
	id, err := tokenPointer.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "the token is malformed")
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "the token doesn't have the subject")
		return
	}

	chunks := strings.Split(bodyJson.Body, " ")
	profanes := []string{"kerfuffle", "sharbert", "fornax"}
	for i, chunk := range chunks {
		for _, profane := range profanes {
			if strings.EqualFold(strings.ToLower(chunk), strings.ToLower(profane)) {
				chunks[i] = "****"
			}
		}
	}

	chirp, err := a.database.CreateChirp(strings.Join(chunks, " "), idInt)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJson(w, 201, chirp)
}

func (a *apiConfig) fetchChirps(w http.ResponseWriter, _ *http.Request) {
	chirps, err := a.database.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	respondWithJson(w, http.StatusOK, chirps)
}
