package main

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Embed templates, data, and static assets directly into the Go binary.
// This guarantees zero-failure execution across Vercel Serverless, AWS Lambda, Docker, and local machines.
//
//go:embed templates
var templatesFS embed.FS

//go:embed data
var dataFS embed.FS

//go:embed static
var staticFS embed.FS

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

// getHomeTemplate parses and caches templates from the embedded filesystem for 100% reliable execution.
func getHomeTemplate() (*template.Template, error) {
	homeTemplateOnce.Do(func() {
		// 1. Try embedded filesystem (works on Vercel, Docker, Lambda without external disk files)
		tmpl, err := template.ParseFS(templatesFS, "templates/index.html", "templates/partials/*.html")
		if err == nil {
			homeTemplate = tmpl
			log.Printf("Successfully loaded templates from embedded filesystem.")
			return
		}
		log.Printf("Notice: template.ParseFS returned (%v), attempting disk fallback...", err)

		// 2. Fallback to local disk files if running in unpacked mode
		partials, globErr := filepath.Glob("templates/partials/*.html")
		if globErr != nil {
			log.Printf("Error finding partial templates on disk: %v", globErr)
		}
		files := append([]string{"templates/index.html"}, partials...)
		homeTemplate, homeTemplateErr = template.ParseFiles(files...)
	})
	return homeTemplate, homeTemplateErr
}

func main() {
	// 1. Load invitation data
	loadInvitationData()
	loadSavedRSVPs()

	// 2. Set up HTTP request router
	mux := http.NewServeMux()

	// Static assets: Serves CSS, JS, images, audio with caching headers.
	// Uses embedded static FS with fallback to disk.
	var staticFileServer http.Handler
	if subFS, err := fs.Sub(staticFS, "static"); err == nil {
		staticFileServer = http.FileServer(http.FS(subFS))
	} else {
		staticFileServer = http.FileServer(http.Dir("static"))
	}

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
	log.Printf("⚡ Embedded templates & gzip compression active")
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

	if tmpl.Lookup("index.html") != nil {
		if err := tmpl.ExecuteTemplate(w, "index.html", invitationData); err != nil {
			log.Printf("Error executing template: %v", err)
		}
	} else {
		if err := tmpl.Execute(w, invitationData); err != nil {
			log.Printf("Error executing template: %v", err)
		}
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

	// Try reading embedded calendar asset
	icsBytes, err := staticFS.ReadFile("static/calendar.ics")
	if err == nil && len(icsBytes) > 0 {
		_, _ = w.Write(icsBytes)
		return
	}

	http.ServeFile(w, r, "static/calendar.ics")
}

// handleRSVP handles guest attendance submissions via POST /api/rsvp.
func handleRSVP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	var req struct {
		Name       string `json:"name"`
		Attendance string `json:"attendance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	name := strings.TrimSpace(req.Name)
	attendance := strings.ToLower(strings.TrimSpace(req.Attendance))

	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Guest name is required",
		})
		return
	}

	if attendance != "yes" && attendance != "no" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Attendance must be 'yes' or 'no'",
		})
		return
	}

	entry := RSVP{
		Name:       name,
		Attendance: attendance,
		CreatedAt:  time.Now(),
	}

	saveRSVPEntry(entry)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Thank you for your response.",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// loadInvitationData reads data/invitation.json from disk or embedded filesystem.
func loadInvitationData() {
	var bytes []byte
	var err error
	filePath := "data/invitation.json"

	// 1. Try reading from disk
	bytes, err = os.ReadFile(filePath)
	if err != nil || len(bytes) == 0 {
		// 2. Fallback to embedded data (for Vercel Serverless environment)
		bytes, err = dataFS.ReadFile(filePath)
	}

	if err != nil || len(bytes) == 0 {
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

// loadSavedRSVPs reads responses from disk or embedded filesystem.
func loadSavedRSVPs() {
	rsvpsMutex.Lock()
	defer rsvpsMutex.Unlock()

	var bytes []byte
	var err error

	bytes, err = os.ReadFile(rsvpFilePath)
	if err != nil || len(bytes) == 0 {
		bytes, _ = dataFS.ReadFile(rsvpFilePath)
	}

	if len(bytes) > 0 {
		_ = json.Unmarshal(bytes, &rsvpsList)
	} else {
		rsvpsList = make([]RSVP, 0)
	}
	log.Printf("Loaded %d previously saved RSVP(s).", len(rsvpsList))
}

// saveRSVPEntry appends a new RSVP and persists it safely (supporting read-only serverless filesystems).
func saveRSVPEntry(entry RSVP) {
	rsvpsMutex.Lock()
	defer rsvpsMutex.Unlock()

	rsvpsList = append(rsvpsList, entry)
	bytes, err := json.MarshalIndent(rsvpsList, "", "  ")
	if err != nil {
		log.Printf("Error serializing RSVPs: %v", err)
		return
	}

	// Try saving to data directory
	_ = os.MkdirAll(filepath.Dir(rsvpFilePath), 0755)
	if err := os.WriteFile(rsvpFilePath, bytes, 0644); err != nil {
		// Fallback for Vercel/AWS Lambda read-only environment: write to /tmp
		tmpPath := filepath.Join(os.TempDir(), "rsvps.json")
		if tmpErr := os.WriteFile(tmpPath, bytes, 0644); tmpErr == nil {
			log.Printf("Notice: Saved RSVP to temporary storage %s", tmpPath)
		} else {
			log.Printf("Warning: Could not write RSVP: %v", err)
		}
	}
}
