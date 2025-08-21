package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rfwlab/rfw/cmd/rfw/build"
	"github.com/rfwlab/rfw/v1/ssr"
)

func main() {
	if err := os.Chdir(".."); err != nil {
		log.Fatalf("chdir failed: %v", err)
	}

	if err := build.Build(nil); err != nil {
		log.Fatalf("build failed: %v", err)
	}

	clientDir := filepath.Join("dist", "client")
	http.Handle("/dist/client/", http.StripPrefix("/dist/client/", http.FileServer(http.Dir(clientDir))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		rendered, err := ssr.RenderFile("index.rtml", map[string]any{"name": name, "count": 0})
		if err != nil {
			http.Error(w, "render failed", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"UTF-8\">\n<title>SSR Example</title>\n</head>\n<body>\n<div id=\"app\" data-hydrate>%s</div>\n<script src=\"/dist/client/wasm_exec.js\"></script>\n<script>\nconst go = new Go();\nWebAssembly.instantiateStreaming(fetch(\"/dist/client/app.wasm?\" + Date.now()), go.importObject).then((result) => { go.run(result.instance); });\n</script>\n</body>\n</html>", rendered)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
