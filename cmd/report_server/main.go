// Command report_server serves the saved trending reports as JSON for the web UI,
// and the built Angular app itself when web/dist exists.
//
//	go run ./cmd/report_server            # API on :8080
//	cd web && npm start                   # UI on :4200, proxies /api to :8080
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"trend_analyzer/src/config"
	"trend_analyzer/src/services"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	webDir := flag.String("web", "web/dist/web/browser", "built Angular app to serve (skipped if missing)")
	flag.Parse()

	reportsDir := config.Load().ReportsDir
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/reports", func(w http.ResponseWriter, r *http.Request) {
		list, err := services.ListReports(reportsDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})

	mux.HandleFunc("GET /api/reports/{name}", func(w http.ResponseWriter, r *http.Request) {
		b, err := services.ReadReport(reportsDir, r.PathValue("name"))
		switch {
		case errors.Is(err, services.ErrReportNotFound):
			http.NotFound(w, r)
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
		}
	})

	if info, err := os.Stat(*webDir); err == nil && info.IsDir() {
		mux.Handle("GET /", spa(*webDir))
		log.Printf("serving UI from %s", *webDir)
	}

	log.Printf("reports from %s, listening on http://localhost%s", reportsDir, *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// spa serves static files and falls back to index.html so Angular routes work on reload.
func spa(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(filepath.Join(dir, filepath.Clean("/"+r.URL.Path))); err != nil {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
}
