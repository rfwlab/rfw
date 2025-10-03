package server

import (
	"bufio"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rfwlab/rfw/cmd/rfw/build"
	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/cmd/rfw/utils"
	hostpkg "github.com/rfwlab/rfw/v1/host"
)

var rebuilds = expvar.NewInt("rebuilds")

type Server struct {
	Port        string
	Host        bool
	Debug       bool
	stopCh      chan os.Signal
	watcher     *fsnotify.Watcher
	hostCmd     *exec.Cmd
	buildType   string
	hostPort    string
	ignoreUntil time.Time
	hmrMu       sync.Mutex
	hmrClients  map[chan []byte]struct{}
}

const ignoreDelay = 200 * time.Millisecond

func NewServer(port string, host, debug bool) *Server {
	utils.EnableDebug(debug)
	return &Server{
		Port:       port,
		Host:       host,
		Debug:      debug,
		stopCh:     make(chan os.Signal, 1),
		hmrClients: make(map[chan []byte]struct{}),
	}
}

func (s *Server) Start() error {
	if err := build.Build(); err != nil {
		return err
	}

	// Detect build type from manifest to know if host components are enabled.
	s.buildType = readBuildType()
	var mux *http.ServeMux
	httpsPort := incrementPort(s.Port)
	if s.buildType == "ssc" {
		s.hostPort = incrementPort(httpsPort)
		if err := s.startHost(); err != nil {
			return err
		}
		target := &url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%s", s.hostPort)}
		proxy := httputil.NewSingleHostReverseProxy(target)
		mux = http.NewServeMux()
		mux.HandleFunc("/__rfw/hmr", s.handleHMR)
		mux.Handle("/", proxy)
		go func() {
			if err := http.ListenAndServe(":"+s.Port, mux); err != nil {
				utils.Fatal("Server failed: ", err)
			}
		}()
		go func() {
			if err := hostpkg.ListenAndServeTLSWithMux(":"+httpsPort, mux); err != nil {
				utils.Fatal("HTTPS server failed: ", err)
			}
		}()
	} else {
		root := filepath.Join("build", "client")
		mux = hostpkg.NewMux(root)
		mux.HandleFunc("/__rfw/hmr", s.handleHMR)
		go func() {
			if err := http.ListenAndServe(":"+s.Port, mux); err != nil {
				utils.Fatal("Server failed: ", err)
			}
		}()
		go func() {
			if err := hostpkg.ListenAndServeTLSWithMux(":"+httpsPort, mux); err != nil {
				utils.Fatal("HTTPS server failed: ", err)
			}
		}()
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = watcher
	if err := s.addWatchers("."); err != nil {
		return err
	}
	go s.watchFiles()

	signal.Notify(s.stopCh, syscall.SIGINT, syscall.SIGTERM)

	localIP := ""
	if s.Host {
		localIP, err = utils.GetLocalIP()
		if err != nil {
			return err
		}
	}

	utils.ClearScreen()
	utils.PrintStartupInfo(s.Port, httpsPort, localIP, s.Host)

	go s.listenForCommands()

	<-s.stopCh
	utils.Info("Server stopped.")
	s.stopHost()
	return nil
}

func (s *Server) listenForCommands() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch strings.ToLower(input) {
		case "h":
			utils.PrintHelp()
		case "u":
			utils.ClearScreen()
			localIP, err := utils.GetLocalIP()
			if err != nil {
				utils.Fatal("Failed to get local IP address: ", err)
			}
			httpsPort := incrementPort(s.Port)
			utils.PrintStartupInfo(s.Port, httpsPort, localIP, s.Host)
		case "c", "q":
			utils.Info("Closing the server...")
			s.stopCh <- syscall.SIGINT
			return
		case "o":
			utils.Info("Opening the browser...")
			url := fmt.Sprintf("http://localhost:%s/", s.Port)
			if err := utils.OpenBrowser(url); err != nil {
				utils.Info(fmt.Sprintf("Failed to open browser: %v", err))
			}
		default:
			utils.Info("Unknown command. Press 'h' for help.")
		}
	}
}

func (s *Server) addWatchers(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name != "." && (name == "build" || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			return s.watcher.Add(path)
		}
		return nil
	})
}

func isGenerated(path string) bool {
	return strings.HasPrefix(filepath.Base(path), "rfw_")
}

func (s *Server) shouldIgnore(now time.Time) bool {
	return now.Before(s.ignoreUntil)
}

func drainWatcher(w *fsnotify.Watcher) {
	for {
		select {
		case <-w.Events:
		case <-w.Errors:
		case <-time.After(50 * time.Millisecond):
			return
		}
	}
}

func (s *Server) watchFiles() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if s.shouldIgnore(time.Now()) {
				continue
			}
			utils.Debug(fmt.Sprintf("event: %s", event))
			if event.Op&fsnotify.Create != 0 {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					name := filepath.Base(event.Name)
					if name == "build" || strings.HasPrefix(name, ".") {
						continue
					}
					utils.Debug(fmt.Sprintf("watching new directory: %s", event.Name))
					_ = s.watcher.Add(event.Name)
					continue
				}
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if isGenerated(event.Name) {
					continue
				}
				if strings.HasSuffix(event.Name, ".go") ||
					strings.HasSuffix(event.Name, ".rtml") ||
					strings.HasSuffix(event.Name, ".md") ||
					plugins.NeedsRebuild(event.Name) {
					rebuilds.Add(1)
					utils.Info("Changes detected, rebuilding...")
					if err := build.Build(); err != nil {
						utils.Fatal("Failed to rebuild project: ", err)
					}
					if strings.HasSuffix(event.Name, ".rtml") {
						markup, err := os.ReadFile(event.Name)
						if err == nil {
							if comps := componentNamesForTemplate(event.Name); len(comps) > 0 {
								for _, name := range comps {
									if err := s.broadcastTemplateUpdate(event.Name, name, string(markup)); err != nil {
										utils.Debug(fmt.Sprintf("template broadcast failed: %v", err))
									}
								}
							} else {
								if err := s.broadcastReload(event.Name); err != nil {
									utils.Debug(fmt.Sprintf("hmr broadcast skipped: %v", err))
								}
							}
						} else {
							utils.Debug(fmt.Sprintf("failed reading template %s: %v", event.Name, err))
							if err := s.broadcastReload(event.Name); err != nil {
								utils.Debug(fmt.Sprintf("hmr broadcast skipped: %v", err))
							}
						}
					} else {
						if err := s.broadcastReload(event.Name); err != nil {
							utils.Debug(fmt.Sprintf("hmr broadcast skipped: %v", err))
						}
					}
					if s.buildType == "ssc" {
						s.stopHost()
						if err := s.startHost(); err != nil {
							utils.Fatal("Failed to restart host server: ", err)
						}
					}
					s.ignoreUntil = time.Now().Add(ignoreDelay)
					drainWatcher(s.watcher)
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			utils.Info(fmt.Sprintf("Watcher error: %v", err))
		case <-s.stopCh:
			s.watcher.Close()
			s.stopHost()
			return
		}
	}
}

// readBuildType reads the build type from rfw.json if present.
func readBuildType() string {
	var manifest struct {
		Build struct {
			Type string `json:"type"`
		} `json:"build"`
	}
	data, err := os.ReadFile("rfw.json")
	if err != nil {
		return ""
	}
	_ = json.Unmarshal(data, &manifest)
	return strings.ToLower(manifest.Build.Type)
}

func incrementPort(port string) string {
	p, _ := strconv.Atoi(port)
	return strconv.Itoa(p + 1)
}

func (s *Server) startHost() error {
	path := filepath.Join("build", "host", "host")
	s.hostCmd = exec.Command(path)
	s.hostCmd.Stdout = os.Stdout
	s.hostCmd.Stderr = os.Stderr
	if s.hostPort != "" {
		env := os.Environ()
		env = append(env, fmt.Sprintf("RFW_HOST_PORT=%s", s.hostPort))
		s.hostCmd.Env = env
	}
	return s.hostCmd.Start()
}

func (s *Server) stopHost() {
	if s.hostCmd != nil && s.hostCmd.Process != nil {
		_ = s.hostCmd.Process.Kill()
		_, _ = s.hostCmd.Process.Wait()
	}
	s.hostCmd = nil
}
