package handlers

import (
	"chirpy/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (a *ApiConfig) generateAccessToken(w http.ResponseWriter, r *http.Request) {
	type ResponseBody struct {
		Token string `json:"token"`
	}
	authorizationToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	user, err := a.Database.GetUserByRefreshToken(string(authorizationToken))
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "something went wrong in the database")
		return
	}
	if len(user.Email) == 0 {
		utils.RespondWithError(w, http.StatusUnauthorized, "this user couldn't be authorized")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(time.Hour))),
		Subject:   fmt.Sprint(user.Id),
	})
	tokenString, err := token.SignedString([]byte(a.JwtSecret))
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := ResponseBody{
		Token: tokenString,
	}
	utils.RespondWithJson(w, http.StatusOK, response)
}

func (a *ApiConfig) revokeUser(w http.ResponseWriter, r *http.Request) {
	authorizationToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	err := a.Database.RevokeRefreshToken(string(authorizationToken))
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "something went wrong in the database")
		return
	}

	utils.RespondWithJson(w, http.StatusNoContent, nil)
}

// Create a helper function to parse the JWT
func (a *ApiConfig) parseJWT(r *http.Request) (*jwt.RegisteredClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}
	tokenString := strings.Split(authHeader, " ")[1]
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(a.JwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
