//go:build !js

// Package build compiles RFW applications and runs build plugins. This file
// covers the WebAssembly artifacts a build produces and how the page that
// boots them is versioned.
package build

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/andybalholm/brotli"
)

// Encoding names an HTTP content coding the build can produce for the wasm
// bundle, in the order a client should prefer them.
type Encoding string

const (
	// EncodingBrotli is the smallest artifact but can only be decoded by the
	// browser's own HTTP layer: no JavaScript DecompressionStream implements
	// brotli, so it is usable only when the server labels it.
	EncodingBrotli Encoding = "br"
	// EncodingGzip is larger but a client can decode it itself through
	// DecompressionStream, which is what makes plain static hosting work.
	EncodingGzip Encoding = "gzip"
)

// artifactExt maps an encoding to the suffix its file carries.
var artifactExt = map[Encoding]string{
	EncodingBrotli: ".br",
	EncodingGzip:   ".gz",
}

// compressors builds the writer for each encoding.
var compressors = map[Encoding]func(io.Writer) io.WriteCloser{
	EncodingBrotli: func(w io.Writer) io.WriteCloser {
		return brotli.NewWriterLevel(w, brotli.BestCompression)
	},
	EncodingGzip: func(w io.Writer) io.WriteCloser {
		writer, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			// BestCompression is a valid level, so this cannot happen.
			return gzip.NewWriter(w)
		}
		return writer
	},
}

// prepareWasmArtifacts writes the artifacts a build will advertise and reports
// the encodings that exist on disk afterwards. Compression is the production
// default; a dev build never recompresses, and an embedded build has no
// network transfer to compress. Both of those cases drop artifacts an earlier
// build left behind, because the loader prefers a compressed artifact over the
// raw one and a stale copy would be served under an immutable cache header.
func prepareWasmArtifacts(wasmPath string, shape buildShape, isDev bool) ([]Encoding, error) {
	if isDev || !shape.compresses() {
		if err := removeStaleArtifacts(wasmPath); err != nil {
			return nil, fmt.Errorf("failed to remove stale wasm artifacts: %w", err)
		}
		return nil, nil
	}
	encodings, err := compressWasm(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compress wasm: %w", err)
	}
	return encodings, nil
}

// compressWasm writes every production artifact next to src and reports the
// encodings that exist afterwards. A build that produces no compressed
// artifact is reported as such rather than being papered over, because the
// client configuration has to describe what is actually on disk.
func compressWasm(src string) ([]Encoding, error) {
	produced := make([]Encoding, 0, len(artifactExt))
	for _, encoding := range []Encoding{EncodingBrotli, EncodingGzip} {
		if err := compressWasmAs(src, encoding); err != nil {
			return nil, err
		}
		produced = append(produced, encoding)
	}
	return produced, nil
}

// compressWasmAs writes one compressed copy of src.
func compressWasmAs(src string, encoding Encoding) (err error) {
	srcRoot, in, err := openFile(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); err == nil {
			err = closeErr
		}
		if closeErr := srcRoot.Close(); err == nil {
			err = closeErr
		}
	}()

	dst, err := projectPath(src + artifactExt[encoding])
	if err != nil {
		return err
	}
	outputRoot, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := outputRoot.Close(); err == nil {
			err = closeErr
		}
	}()
	if err := outputRoot.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	out, err := outputRoot.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()

	writer := compressors[encoding](out)
	if _, err := io.Copy(writer, in); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}
	return writer.Close()
}

// removeStaleArtifacts drops compressed copies left by an earlier build. A dev
// build never recompresses, so a stale artifact would otherwise be served under
// an immutable cache header and shadow the fresh bundle.
func removeStaleArtifacts(src string) (err error) {
	root, err := os.OpenRoot(".")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := root.Close(); err == nil {
			err = closeErr
		}
	}()
	for _, ext := range artifactExt {
		dst, pathErr := projectPath(src + ext)
		if pathErr != nil {
			return pathErr
		}
		if removeErr := root.Remove(dst); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return removeErr
		}
	}
	return nil
}

// versionedScripts are the bootstrap files a release must be able to
// invalidate. A cached loader from a previous release cannot be allowed to
// drive a new bundle, and the configuration cannot version the tag that loads
// it, so the build stamps the markup instead.
var versionedScripts = []string{"rfw_config.js", "wasm_exec.js", "wasm_loader.js"}

// scriptTag matches a bootstrap script tag with an optional existing query.
// The match is deliberately narrow: an unrecognised tag is left untouched
// rather than rewritten, because this runs over application-authored markup.
var scriptTag = regexp.MustCompile(`(<script\b[^>]*\bsrc=")(/?)(` + quotedScripts() + `)(\?[^"]*)?(")`)

// quotedScripts builds the alternation of bootstrap file names with their
// metacharacters escaped, so a dot matches a dot.
func quotedScripts() string {
	quoted := make([]string, len(versionedScripts))
	for i, name := range versionedScripts {
		quoted[i] = regexp.QuoteMeta(name)
	}
	return strings.Join(quoted, "|")
}

// stampBootstrapScripts rewrites the bootstrap script tags in html so each
// carries the current build version. It reports whether anything changed.
func stampBootstrapScripts(html, version string) (string, bool) {
	stamped := scriptTag.ReplaceAllString(html, `${1}${2}${3}?v=`+version+`${5}`)
	return stamped, stamped != html
}

// stampIndexHTML applies stampBootstrapScripts to the built index.html.
func stampIndexHTML(path, version string) error {
	data, err := readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	stamped, changed := stampBootstrapScripts(string(data), version)
	if !changed {
		return nil
	}
	return writePublicFile(path, []byte(stamped))
}
