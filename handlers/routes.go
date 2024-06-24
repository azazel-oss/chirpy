package handlers

import (
	"chirpy/internal/database"
	"net/http"
)

type ApiConfig struct {
	Database       *database.DB
	JwtSecret      string
	FileserverHits int
}

func RegisterRoutes(mux *http.ServeMux, apiCfg *ApiConfig) {
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/admin/metrics", apiCfg.handleMetricsEndpoint)
	mux.HandleFunc("/api/reset", apiCfg.handleResetEndpoint)
	mux.HandleFunc("/api/healthz", handleReadinessEndpoint)

	mux.HandleFunc("GET /api/chirps", apiCfg.fetchChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.fetchSingleChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirps)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", apiCfg.deleteSingleChirp)

	mux.HandleFunc("POST /api/users", apiCfg.createUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUser)
	mux.HandleFunc("POST /api/login", apiCfg.loginUser)

	mux.HandleFunc("POST /api/refresh", apiCfg.generateAccessToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeUser)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.polkaUpgradeHandler)
}
