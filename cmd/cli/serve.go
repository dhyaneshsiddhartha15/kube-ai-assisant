package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// servePort is the port for the HTTP server (set by serve command).
var servePort int

// request body for POST /generate
type generateRequest struct {
	Prompt string `json:"prompt"`
}

// response for POST /generate
type generateResponse struct {
	YAML string `json:"yaml"`
}

// errResponse for error responses
type errResponse struct {
	Error string `json:"error"`
}

func writeCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleGenerate(oaic *oaiClients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(errResponse{Error: "method not allowed, use POST"})
			return
		}

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResponse{Error: "invalid JSON or missing prompt: " + err.Error()})
			return
		}
		if req.Prompt == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResponse{Error: "prompt is required"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
		defer cancel()

		prompts := []string{req.Prompt}
		var yaml string
		var err error
		if useGemini() {
			yaml, err = geminiCompletion(ctx, prompts, *geminiModel)
		} else {
			yaml, err = gptCompletion(ctx, *oaic, prompts, *openAIDeploymentName)
		}
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errResponse{Error: err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(generateResponse{YAML: yaml})
	}
}

// runServe starts the HTTP server for the generate API.
func runServe() error {
	var oaic oaiClients
	if !useGemini() {
		var err error
		oaic, err = newOAIClients()
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/generate", handleGenerate(&oaic))

	addr := ":" + strconv.Itoa(servePort)
	log.Infof("Listening on http://localhost%s - POST /generate with {\"prompt\": \"...\"}", addr)
	return http.ListenAndServe(addr, mux)
}

func setServePort(port int) { servePort = port }
