package main

import (
	"compress/gzip"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RSVP represents a guest's attendance response.
type RSVP struct {
	Name       string    `json:"name"`
	Attendance string    `json:"attendance"` // "yes" or "no"
	CreatedAt  time.Time `json:"created_at"`
}

// Global state for in-memory RSVP storage and wedding invitation data.
var (
	invitationData   map[string]interface{}
	rsvpsMutex       sync.Mutex
	rsvpsList        []RSVP
	rsvpFilePath     = "data/rsvps.json"
	homeTemplate     *template.Template
	homeTemplateOnce sync.Once
	homeTemplateErr  error
)

// gzipResponseWriter wraps http.ResponseWriter to write gzip compressed data.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// withGzip compresses HTTP responses for clients supporting gzip.
func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		// Skip compression for binaries and pre-compressed media
		path := r.URL.Path
		if strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".mp3") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzWriter := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzWriter, r)
	})
}

// getHomeTemplate parses and caches templates in memory for zero disk I/O latency.
func getHomeTemplate() (*template.Template, error) {
	homeTemplateOnce.Do(func() {
		partials, err := filepath.Glob("templates/partials/*.html")
		if err != nil {
			log.Printf("Error finding partial templates: %v", err)
		}
		files := append([]string{"templates/index.html"}, partials...)
		homeTemplate, homeTemplateErr = template.ParseFiles(files...)
	})
	return homeTemplate, homeTemplateErr
}

func main() {
	// 1. Load invitation data from JSON file
	loadInvitationData()
	loadSavedRSVPs()

	// 2. Set up HTTP request router
	mux := http.NewServeMux()

	// Static assets: Serves CSS, JS, images, audio with caching headers
	staticFileServer := http.FileServer(http.Dir("static"))
	staticHandler := http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") {
			w.Header().Set("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")
		} else if strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".webp") || strings.HasSuffix(path, ".mp3") {
			w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		staticFileServer.ServeHTTP(w, r)
	}))
	mux.Handle("/static/", staticHandler)

	// Routes
	mux.HandleFunc("/", handleHome)                 // GET / (Serves HTML templates)
	mux.HandleFunc("/health", handleHealth)         // GET /health (Health check)
	mux.HandleFunc("/calendar.ics", handleCalendar) // GET /calendar.ics (iCalendar file)
	mux.HandleFunc("/api/rsvp", handleRSVP)         // POST /api/rsvp (Attendance Submission)

	// Port configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := ":" + port
	log.Printf("==================================================")
	log.Printf("💍 Wedding Invitation Server is running")
	log.Printf("🌐 Local URL: http://localhost:%s", port)
	log.Printf("📄 Routes:")
	log.Printf("   - GET  /         (Invitation Website)")
	log.Printf("   - GET  /health   (Health Status)")
	log.Printf("   - POST /api/rsvp (Attendance Submission)")
	log.Printf("⚡ Gzip compression & asset caching enabled")
	log.Printf("==================================================")

	if err := http.ListenAndServe(serverAddr, withGzip(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// handleHome serves the HTML templates for the wedding invitation.
func handleHome(w http.ResponseWriter, r *http.Request) {
	// Only serve root path; return 404 for unknown routes
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := getHomeTemplate()
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, "Internal Server Error: Unable to parse templates", http.StatusInternalServerError)
		return
	}

	// Render the invitation template with invitation data
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if err := tmpl.Execute(w, invitationData); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// handleHealth returns a simple JSON response indicating the server is operational.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Wedding invitation service is healthy",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handleCalendar serves an iCalendar (.ics) file for the wedding event.
func handleCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"wedding-fahad-and-samar.ics\"")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	http.ServeFile(w, r, "static/calendar.ics")
}

// handleRSVP handles guest attendance submissions via POST /api/rsvp.
//
// Expected Request JSON:
//
//	{
//	  "name": "Guest Name",
//	  "attendance": "yes" // or "no"
//	}
//
// Returns HTTP 200 on success:
//
//	{
//	  "success": true,
//	  "message": "Thank you for your response."
//	}
func handleRSVP(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// Request payload struct
	var req struct {
		Name       string `json:"name"`
		Attendance string `json:"attendance"`
	}

	// Decode incoming JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate: name is required
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "name is required",
		})
		return
	}

	// Validate: attendance must be "yes" or "no"
	attendance := strings.ToLower(strings.TrimSpace(req.Attendance))
	if attendance != "yes" && attendance != "no" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "attendance must be yes or no",
		})
		return
	}

	// Record RSVP entry
	entry := RSVP{
		Name:       name,
		Attendance: attendance,
		CreatedAt:  time.Now(),
	}

	saveRSVPEntry(entry)

	// Return successful JSON response
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Thank you for your response.",
	})
}

// writeJSON writes a standardized JSON response with the given status code.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// loadInvitationData reads data/invitation.json into memory.
func loadInvitationData() {
	filePath := "data/invitation.json"
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Notice: Could not read %s (%v). Using empty invitation data.", filePath, err)
		invitationData = make(map[string]interface{})
		return
	}

	if err := json.Unmarshal(bytes, &invitationData); err != nil {
		log.Printf("Warning: Failed to parse %s: %v", filePath, err)
		invitationData = make(map[string]interface{})
		return
	}

	log.Printf("Loaded invitation metadata successfully.")
}

// loadSavedRSVPs reads previously saved responses from data/rsvps.json.
func loadSavedRSVPs() {
	rsvpsMutex.Lock()
	defer rsvpsMutex.Unlock()

	bytes, err := os.ReadFile(rsvpFilePath)
	if err != nil {
		rsvpsList = make([]RSVP, 0)
		return
	}

	_ = json.Unmarshal(bytes, &rsvpsList)
	log.Printf("Loaded %d previously saved RSVP(s).", len(rsvpsList))
}

// saveRSVPEntry appends a new RSVP to memory and persists it to data/rsvps.json.
func saveRSVPEntry(entry RSVP) {
	rsvpsMutex.Lock()
	defer rsvpsMutex.Unlock()

	rsvpsList = append(rsvpsList, entry)
	bytes, err := json.MarshalIndent(rsvpsList, "", "  ")
	if err != nil {
		log.Printf("Error serializing RSVPs: %v", err)
		return
	}

	// Ensure data directory exists
	_ = os.MkdirAll(filepath.Dir(rsvpFilePath), 0755)
	if err := os.WriteFile(rsvpFilePath, bytes, 0644); err != nil {
		log.Printf("Warning: could not write %s: %v", rsvpFilePath, err)
	}
}
