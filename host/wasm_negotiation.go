package host

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// wasmEncoding pairs a content coding with the suffix of the artifact that
// carries it. The order is the server's preference, best first.
type wasmEncoding struct {
	name string
	ext  string
}

var wasmEncodings = []wasmEncoding{
	{name: "br", ext: ".br"},
	{name: "gzip", ext: ".gz"},
}

// negotiateWasm answers a request for the raw bundle with the best encoding
// the client accepts and the build actually produced. Serving the compressed
// artifact under the raw URL is what lets one URL work for every client: the
// browser advertises what it can decode, and no application configuration has
// to guess.
//
// It reports whether it wrote the response.
func negotiateWasm(w http.ResponseWriter, r *http.Request, dir http.Dir, urlPath string) bool {
	if !strings.HasSuffix(urlPath, ".wasm") {
		return false
	}
	accepted := acceptedEncodings(r.Header.Get("Accept-Encoding"))
	if len(accepted) == 0 {
		return false
	}
	for _, encoding := range wasmEncodings {
		if !accepted[encoding.name] {
			continue
		}
		file, size, ok := openArtifact(dir, urlPath+encoding.ext)
		if !ok {
			continue
		}
		defer func() {
			_ = file.Close()
		}()
		header := w.Header()
		header.Set("Content-Type", "application/wasm")
		header.Set("Content-Encoding", encoding.name)
		// The length of what actually goes on the wire. A client that measures
		// progress against it is measuring compressed bytes, which is why the
		// loader treats an encoded response as indeterminate.
		header.Set("Content-Length", strconv.FormatInt(size, 10))
		addVaryAcceptEncoding(header)
		setWasmCacheControl(header, r.URL.Query().Get("v") != "")
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return true
		}
		if _, err := io.Copy(w, file); err != nil {
			logger.Warn("wasm negotiation write failed", "path", urlPath, "error", err)
		}
		return true
	}
	return false
}

// openArtifact opens a compressed bundle and reports its size on disk.
func openArtifact(dir http.Dir, name string) (http.File, int64, bool) {
	file, err := dir.Open(filepath.ToSlash(name))
	if err != nil {
		return nil, 0, false
	}
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		_ = file.Close()
		return nil, 0, false
	}
	return file, info.Size(), true
}

// acceptedEncodings parses an Accept-Encoding header into the set of codings
// the client will take. A coding with q=0 is an explicit refusal, so it is
// dropped rather than merely deprioritised.
func acceptedEncodings(header string) map[string]bool {
	if strings.TrimSpace(header) == "" {
		return nil
	}
	accepted := map[string]bool{}
	// A refusal is remembered separately: "*, br;q=0" accepts anything except
	// brotli, so expanding the wildcard must not put brotli back.
	refused := map[string]bool{}
	for _, part := range strings.Split(header, ",") {
		name, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			continue
		}
		if quality, ok := parseQuality(params); ok && quality == 0 {
			refused[name] = true
			delete(accepted, name)
			continue
		}
		accepted[name] = true
	}
	if accepted["*"] {
		for _, encoding := range wasmEncodings {
			if !refused[encoding.name] {
				accepted[encoding.name] = true
			}
		}
	}
	return accepted
}

// parseQuality reads the q value out of an Accept-Encoding parameter list.
func parseQuality(params string) (float64, bool) {
	for _, param := range strings.Split(params, ";") {
		key, value, found := strings.Cut(strings.TrimSpace(param), "=")
		if !found || !strings.EqualFold(strings.TrimSpace(key), "q") {
			continue
		}
		quality, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 0, false
		}
		return quality, true
	}
	return 0, false
}

// addVaryAcceptEncoding records that the response body depends on the request
// encoding, so a shared cache cannot hand a brotli body to a client that
// cannot decode it.
func addVaryAcceptEncoding(header http.Header) {
	vary := header.Get("Vary")
	switch {
	case vary == "":
		header.Set("Vary", "Accept-Encoding")
	case !strings.Contains(vary, "Accept-Encoding"):
		header.Set("Vary", vary+", Accept-Encoding")
	}
}

// revalidates reports whether a document has to be checked with the server on
// every load. Both entries here carry pointers to versioned assets:
// index.html holds the stamped bootstrap script tags and rfw_config.js holds
// the wasm version. Caching either one lets a browser keep following an old
// pointer, so a release could never replace the loader or the bundle.
func revalidates(path string) bool {
	switch strings.Trim(path, "/") {
	case "index.html", "rfw_config.js":
		return true
	default:
		return false
	}
}

// setWasmCacheControl applies the caching policy for a bundle. A URL that
// carries a content version is immutable; one that does not has to be
// revalidated or a release cannot replace it.
func setWasmCacheControl(header http.Header, versioned bool) {
	if devMode() {
		header.Set("Cache-Control", "no-store")
		return
	}
	if versioned {
		header.Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	header.Set("Cache-Control", "no-cache")
}

// artifactEncoding reports the coding a directly requested artifact carries.
func artifactEncoding(path string) (string, bool) {
	if !strings.Contains(path, ".wasm") {
		return "", false
	}
	for _, encoding := range wasmEncodings {
		if strings.HasSuffix(path, ".wasm"+encoding.ext) {
			return encoding.name, true
		}
	}
	return "", false
}

// wasmRequest reports whether path addresses a bundle in any encoding.
func wasmRequest(path string) bool {
	if strings.HasSuffix(path, ".wasm") {
		return true
	}
	_, ok := artifactEncoding(path)
	return ok
}
