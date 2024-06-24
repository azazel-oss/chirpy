package handlers

import (
	"chirpy/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (a *ApiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	type ResponseBody struct {
		Email        string `json:"email"`
		Id           int    `json:"id"`
		IsChirpyUser bool   `json:"is_chirpy_red"`
	}
	type RequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}
	claims, err := a.parseJWT(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "the user is not authorized")
		return
	}
	id, err := claims.GetSubject()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "the token is malformed")
		return
	}
	user, err := a.Database.UpdateUser(id, bodyJson.Email, bodyJson.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "we couldn't update the user")
		return
	}

	response := ResponseBody{
		Email:        user.Email,
		Id:           user.Id,
		IsChirpyUser: user.IsRedUser,
	}
	utils.RespondWithJson(w, http.StatusOK, response)
}

func (a *ApiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
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
		IsRedUser    bool   `json:"is_chirpy_red"`
	}
	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}
	user, err := a.Database.LoginUser(bodyJson.Email, bodyJson.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
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
	tokenString, err := token.SignedString([]byte(a.JwtSecret))
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := ResponseBody{
		Email:        user.Email,
		Id:           user.Id,
		Token:        tokenString,
		RefreshToken: user.RefreshToken,
		IsRedUser:    user.IsRedUser,
	}
	utils.RespondWithJson(w, http.StatusOK, response)
}

func (a *ApiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type ResponseBody struct {
		Email        string `json:"email"`
		Id           int    `json:"id"`
		IsChirpyUser bool   `json:"is_chirpy_red"`
	}

	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}

	user, err := a.Database.CreateUser(bodyJson.Email, bodyJson.Password)
	response := ResponseBody{
		Email:        user.Email,
		Id:           user.Id,
		IsChirpyUser: user.IsRedUser,
	}
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondWithJson(w, http.StatusCreated, response)
}
