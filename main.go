package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	//	"github.com/Serux/chirpy/internal/auth"
	"github.com/Serux/chirpy/internal/auth"
	"github.com/Serux/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	jwtSecret      string
	queries        *database.Queries
}

type fullChirpJsonDb struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}

type userMailJsonDb struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}
func (cfg *apiConfig) metricsHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	template := `
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`
	val := fmt.Sprintf(template, cfg.fileserverHits.Load())
	rw.Write([]byte(val))
}
func (cfg *apiConfig) resetHandler(rw http.ResponseWriter, r *http.Request) {
	if os.Getenv("PLATFORM") != "dev" {
		respondWithError(rw, http.StatusForbidden, "FORBIDDEN")
		return
	}
	err := cfg.queries.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "ERROR DELETING USERS")
		return
	}

	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	val := "Hits Reset."
	rw.Write([]byte(val))
}
func (cfg *apiConfig) postUsersHandler(rw http.ResponseWriter, r *http.Request) {
	type requestJson struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestJson{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong decoding input")
		return
	}

	user, err := cfg.queries.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: params.Password})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong creating user")
		return
	}

	ret := userMailJsonDb{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondWithJSON(rw, http.StatusCreated, ret)
}

func (cfg *apiConfig) putUsersHandler(rw http.ResponseWriter, r *http.Request) {
	type requestJson struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Something went wrong geting JWT")
		return
	}
	uidtok, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Something went wrong validating JWT")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestJson{}

	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong decoding input")
		return
	}

	user, err := cfg.queries.UpdateUserMailPassByUUID(r.Context(), database.UpdateUserMailPassByUUIDParams{Email: params.Email, HashedPassword: params.Password, ID: uidtok})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong creating user")
		return
	}

	ret := userMailJsonDb{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondWithJSON(rw, http.StatusOK, ret)
}

func (cfg *apiConfig) postChirpsHandler(rw http.ResponseWriter, r *http.Request) {
	type requestJson struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Something went wrong geting JWT")
		return
	}
	uidtok, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Something went wrong validating JWT")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestJson{}

	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong decoding input")
		return
	}

	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{Body: params.Body, UserID: uidtok})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong creating user")
		return
	}

	ret := fullChirpJsonDb{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
	}

	respondWithJSON(rw, http.StatusCreated, ret)
}
func (cfg *apiConfig) getChirpsHandler(rw http.ResponseWriter, r *http.Request) {

	chirp, err := cfg.queries.SelectAllChirps(r.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong creating user")
		return
	}

	ret := []fullChirpJsonDb{}

	for _, ch := range chirp {
		ret = append(ret, fullChirpJsonDb{
			Id:        ch.ID.String(),
			CreatedAt: ch.CreatedAt.Format(time.RFC3339),
			UpdatedAt: ch.UpdatedAt.Format(time.RFC3339),
			Body:      ch.Body,
			UserId:    ch.UserID.String(),
		})
	}

	respondWithJSON(rw, http.StatusOK, ret)
}
func (cfg *apiConfig) getChirpHandler(rw http.ResponseWriter, r *http.Request) {

	uid, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		fmt.Println(err)
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong parsing UID")
		return
	}

	ch, err := cfg.queries.SelectOneChirps(r.Context(), uid)
	if err != nil {
		respondWithError(rw, http.StatusNotFound, "Chirp not found")
		return
	}

	ret := fullChirpJsonDb{
		Id:        ch.ID.String(),
		CreatedAt: ch.CreatedAt.Format(time.RFC3339),
		UpdatedAt: ch.UpdatedAt.Format(time.RFC3339),
		Body:      ch.Body,
		UserId:    ch.UserID.String(),
	}

	respondWithJSON(rw, http.StatusOK, ret)
}

func (cfg *apiConfig) deleteChirpHandler(rw http.ResponseWriter, r *http.Request) {

	uid, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		fmt.Println(err)
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong parsing UID")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Something went wrong geting JWT")
		return
	}
	uidtok, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Something went wrong validating JWT")
		return
	}

	ch, err := cfg.queries.SelectOneChirps(r.Context(), uid)
	if err != nil {
		respondWithError(rw, http.StatusNotFound, "Chirp not found")
		return
	}
	if ch.UserID != uidtok {
		respondWithError(rw, 403, "NO AUTH")
		return
	}

	err = cfg.queries.DeleteByIdChirps(r.Context(), database.DeleteByIdChirpsParams{ID: uid, UserID: uidtok})
	if err != nil {
		respondWithError(rw, 403, "Error deleting CHIRP")
		return
	}

	respondWithJSON(rw, 204, nil)
}

func (cfg *apiConfig) loginHandler(rw http.ResponseWriter, r *http.Request) {
	type requestJson struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type responseJson struct {
		Id           string `json:"id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestJson{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong decoding input")
		return
	}

	user, err := cfg.queries.SelectUserByMail(r.Context(), params.Email)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	expires := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expires)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error creating JWT")
		return
	}

	rtoken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error making rtoken"+err.Error())
		return
	}

	_, err = cfg.queries.InsertRefreshToken(r.Context(), database.InsertRefreshTokenParams{Token: rtoken, UserID: user.ID})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error inserting token"+err.Error())
		return
	}

	ret := responseJson{
		Id:           user.ID.String(),
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		Email:        user.Email,
		Token:        token,
		RefreshToken: rtoken,
	}

	respondWithJSON(rw, http.StatusOK, ret)
}

func (cfg *apiConfig) refreshHandler(rw http.ResponseWriter, r *http.Request) {

	type responseJson struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil || token == "" {
		respondWithError(rw, http.StatusUnauthorized, "Header without Bearer Token")
		return
	}

	rt, err := cfg.queries.SelectRefreshToken(r.Context(), token)
	if err != nil || token == "" {
		respondWithError(rw, http.StatusUnauthorized, "No Refresh Token")
		return
	}
	if rt.RevokedAt.Valid || time.Now().After(rt.ExpiresAt.Time) {
		respondWithError(rw, http.StatusUnauthorized, "Token revoked or expired")
		return
	}

	newtoken, err := auth.MakeJWT(rt.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error creating JWT")
		return
	}

	ret := responseJson{
		Token: newtoken,
	}
	respondWithJSON(rw, http.StatusOK, ret)
}

func (cfg *apiConfig) revokeHandler(rw http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil || token == "" {
		respondWithError(rw, http.StatusUnauthorized, "Header without Bearer Token")
		return
	}

	err = cfg.queries.RevokeRefreshToken(r.Context(), token)
	if err != nil || token == "" {
		respondWithError(rw, http.StatusUnauthorized, "No Refresh Token")
		return
	}

	respondWithJSON(rw, 204, nil)
}

func main() {
	fmt.Println("Start Server")
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWTSECRET")

	fmt.Println("Load ENV")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("ERROR OPENING DB", err)
		return
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiConf := apiConfig{}

	apiConf.queries = dbQueries
	apiConf.jwtSecret = jwtSecret

	//APP FILESERVER
	appFileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiConf.middlewareMetricsInc(appFileServerHandler))

	//API
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("POST /api/users", apiConf.postUsersHandler)
	mux.HandleFunc("PUT /api/users", apiConf.putUsersHandler)

	mux.HandleFunc("POST /api/login", apiConf.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiConf.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiConf.revokeHandler)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("POST /api/chirps", apiConf.postChirpsHandler)
	mux.HandleFunc("GET /api/chirps", apiConf.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConf.getChirpHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConf.deleteChirpHandler)

	//ADMIN
	mux.HandleFunc("GET /admin/metrics", apiConf.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConf.resetHandler)

	//START SERVER
	server := http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}

func healthzHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(http.StatusText(http.StatusOK)))
}
func validateChirpHandler(rw http.ResponseWriter, r *http.Request) {
	type requestJson struct {
		Body string `json:"body"`
	}
	type responseJson struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestJson{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(rw, http.StatusBadRequest, "Chirp is too long")
		return
	}

	newBody := removeProfanity(params.Body)

	respondWithJSON(rw, http.StatusOK, responseJson{CleanedBody: newBody})

}
func removeProfanity(chirp string) string {
	profanes := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(chirp)
	for i, w := range words {
		for _, p := range profanes {
			if strings.ToLower(w) == p {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}
func respondWithError(rw http.ResponseWriter, code int, msg string) {
	type errorJson struct {
		Error string `json:"error"`
	}
	errJ, _ := json.Marshal(errorJson{Error: msg})
	rw.WriteHeader(code)
	rw.Write(errJ)
}
func respondWithJSON(rw http.ResponseWriter, code int, payload interface{}) {
	retJ, _ := json.Marshal(payload)
	rw.WriteHeader(code)
	rw.Write(retJ)
}
