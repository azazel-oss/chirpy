package handlers

import (
	"chirpy/utils"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func (a *ApiConfig) polkaUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}
	bodyJson := RequestBody{}
	if len(r.Header.Get("Authorization")) == 0 {
		utils.RespondWithError(w, http.StatusUnauthorized, "You aren't authorized")
		return
	}
	apiToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	if !strings.EqualFold(apiToken, os.Getenv("POLKA_API_KEY")) {
		utils.RespondWithError(w, http.StatusUnauthorized, "You aren't authorized")
		return
	}
	err := json.NewDecoder(r.Body).Decode(&bodyJson)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "couldn't convert body")
		return
	}
	if strings.EqualFold(bodyJson.Event, "user.upgraded") {
		err := a.Database.UpgradeUserToRed(bodyJson.Data.UserID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "")
			return
		}
		utils.RespondWithJson(w, http.StatusNoContent, nil)
	} else {
		utils.RespondWithJson(w, http.StatusNoContent, nil)
	}
}
