package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
	//val := fmt.Sprint("Hits: ", cfg.fileserverHits.Load())
	rw.Write([]byte(val))
}
func (cfg *apiConfig) nRequestResetHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	val := "Hits Reset."
	rw.Write([]byte(val))
}

func main() {
	mux := http.NewServeMux()
	apiConf := apiConfig{}

	//APP FILESERVER
	appFileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiConf.middlewareMetricsInc(appFileServerHandler))

	//API
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	//ADMIN
	mux.HandleFunc("GET /admin/metrics", apiConf.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConf.nRequestResetHandler)

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
		CleanedBody string `json:"cleaned_Body"`
	}
	type errorJson struct {
		Error string `json:"error"`
	}
	decoder := json.NewDecoder(r.Body)
	params := requestJson{}

	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		errJ, _ := json.Marshal(errorJson{Error: "Chirp is too long"})
		json.Marshal(errorJson{Error: "Something went wrong"})
		rw.WriteHeader(500)
		rw.Write(errJ)
		return
	}

	if len(params.Body) > 140 {
		errJ, _ := json.Marshal(errorJson{Error: "Chirp is too long"})
		rw.WriteHeader(400)
		rw.Write(errJ)
		return
	}

	newBody := removeProfanity(params.Body)

	retJ, _ := json.Marshal(responseJson{CleanedBody: newBody})
	rw.WriteHeader(200)
	rw.Write(retJ)

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
