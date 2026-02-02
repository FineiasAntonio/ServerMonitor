package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ServerMonitor/internal/model"
	"ServerMonitor/internal/service"
)

// RegisterRoutes sets up the HTTP handlers
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/host", handleHostInfo)
	mux.HandleFunc("/api/containers", handleContainers)
	mux.HandleFunc("/api/manage", handleContainerAction)
	mux.HandleFunc("/api/containers/", handleContainerRequest)
}

func handleHostInfo(w http.ResponseWriter, r *http.Request) {
	info := service.GetHostMetrics()
	respondJSON(w, info)
}

func handleContainers(w http.ResponseWriter, r *http.Request) {
	list, err := service.ListContainers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, list)
}

func handleContainerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ContainerActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := service.PerformAction(req.ID, req.Action); err != nil {
		http.Error(w, fmt.Sprintf("Error executing %s: %v", req.Action, err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": "success", "action": req.Action, "id": req.ID})
}

func handleContainerRequest(w http.ResponseWriter, r *http.Request) {
	// Path expected: /api/containers/{id}/logs
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) >= 5 && parts[4] == "logs" {
		containerID := parts[3]
		handleContainerLogs(w, r, containerID)
		return
	}

	http.NotFound(w, r)
}

func handleContainerLogs(w http.ResponseWriter, r *http.Request, containerID string) {
	logs, err := service.GetLogs(containerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading logs: %v", err), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.Copy(w, logs)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
