package handlers

import (
	"chirpy/internal/database"
	"chirpy/utils"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (a *ApiConfig) deleteSingleChirp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("chirpId"))
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "provide correct id")
		return
	}
	claims, err := a.parseJWT(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "the user is not authorized")
		return
	}
	userId, err := claims.GetSubject()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "the token is malformed")
	}
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "sorry I messed up")
	}
	err = a.Database.DeleteChirp(id, userIdInt)
	if err != nil {
		utils.RespondWithError(w, http.StatusForbidden, err.Error())
		return
	}
	utils.RespondWithJson(w, http.StatusNoContent, nil)
}

func (a *ApiConfig) fetchSingleChirp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("chirpId"))
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "provide correct id")
		return
	}
	chirp, err := a.Database.GetSingleChirp(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}
	utils.RespondWithJson(w, http.StatusOK, chirp)
}

func (a *ApiConfig) createChirps(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Body string `json:"body"`
	}

	bodyJson := RequestBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(bodyJson.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
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

	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "the token doesn't have the subject")
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

	chirp, err := a.Database.CreateChirp(strings.Join(chunks, " "), idInt)
	if err != nil {
		utils.RespondWithError(w, 500, err.Error())
		return
	}
	utils.RespondWithJson(w, 201, chirp)
}

func (a *ApiConfig) fetchChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := a.Database.GetChirps()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	authorId := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")
	if len(sortQuery) > 0 {
		if sortQuery == "asc" {
			sort.Slice(chirps, func(i, j int) bool {
				return chirps[j].Id > chirps[i].Id
			})
			utils.RespondWithJson(w, http.StatusOK, chirps)
		}
		if sortQuery == "desc" {
			sort.Slice(chirps, func(i, j int) bool {
				return chirps[j].Id < chirps[i].Id
			})
			utils.RespondWithJson(w, http.StatusOK, chirps)
		}
		return
	}
	if len(authorId) > 0 {
		authorIdInt, err := strconv.Atoi(authorId)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "the author id is not well formatted")
			return
		}
		response := []database.Chirp{}
		for _, value := range chirps {
			if value.AuthorId == authorIdInt {
				response = append(response, value)
			}
		}
		utils.RespondWithJson(w, http.StatusOK, response)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, chirps)
}
