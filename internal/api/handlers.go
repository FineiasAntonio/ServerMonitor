package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ServerMonitor/internal/model"
	"ServerMonitor/internal/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for now
	},
}

// RegisterRoutes sets up the HTTP handlers
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/host", handleHostInfo)
	mux.HandleFunc("/api/host/console", handleHostConsole)
	mux.HandleFunc("/api/containers", handleContainers)
	mux.HandleFunc("/api/manage", handleContainerAction)
	mux.HandleFunc("/api/containers/", handleContainerRequest)
}

func handleHostInfo(w http.ResponseWriter, r *http.Request) {
	info := service.GetHostMetrics()
	respondJSON(w, info)
}

func handleHostConsole(w http.ResponseWriter, r *http.Request) {
	// 1. Upgrade HTTP to WS
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// 2. Start Host PTY
	ptmx, err := service.StartHostConsole()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error starting pty: %v", err)))
		return
	}
	defer ptmx.Close()

	// 3. Pipe
	// TextMessage vs BinaryMessage: xterm.js usually handles string data fine.
	// PTY read/write are bytes.

	// WS -> PTY
	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}
			ptmx.Write(msg)
		}
	}()

	// PTY -> WS
	buf := make([]byte, 1024)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			break
		}
		if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
			break
		}
	}
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
	// Path expected: /api/containers/{id}/logs OR /api/containers/{id}/console
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) >= 5 {
		containerID := parts[3]
		action := parts[4]

		if action == "logs" {
			handleContainerLogs(w, r, containerID)
			return
		} else if action == "console" {
			handleContainerConsole(w, r, containerID)
			return
		}
	}

	http.NotFound(w, r)
}

func handleContainerConsole(w http.ResponseWriter, r *http.Request, containerID string) {
	// 1. Upgrade HTTP to WS
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// upgrader.Upgrade handles error response
		return
	}
	defer ws.Close()

	// 2. Create Exec Instance
	execID, err := service.CreateExec(containerID)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error creating exec: %v", err)))
		return
	}

	// 3. Attach and Stream
	if err := service.AttachExec(execID, ws); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error attaching exec: %v", err)))
	}
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
