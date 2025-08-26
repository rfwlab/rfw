package server

import (
	"bufio"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/rfwlab/rfw/cmd/rfw/build"
	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/cmd/rfw/utils"
	hostpkg "github.com/rfwlab/rfw/v1/host"
)

var rebuilds = expvar.NewInt("rebuilds")

type Server struct {
	Port      string
	Host      bool
	Debug     bool
	stopCh    chan os.Signal
	watcher   *fsnotify.Watcher
	hostCmd   *exec.Cmd
	buildType string
}

func NewServer(port string, host, debug bool) *Server {
	utils.EnableDebug(debug)
	return &Server{
		Port:   port,
		Host:   host,
		Debug:  debug,
		stopCh: make(chan os.Signal, 1),
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
		if err := s.startHost(); err != nil {
			return err
		}
	} else {
		mux = hostpkg.NewMux(".")
		if s.Debug {
			mux.Handle("/debug/vars", expvar.Handler())
			mux.HandleFunc("GET /debug/pprof/", pprof.Index)
			mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
		}
		go func() {
			if err := http.ListenAndServe(":"+s.Port, mux); err != nil {
				utils.Fatal("Server failed: ", err)
			}
		}()
		go func() {
			if err := hostpkg.ListenAndServeTLS(":"+httpsPort, "."); err != nil {
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
			utils.OpenBrowser(url)
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
			return s.watcher.Add(path)
		}
		return nil
	})
}

func (s *Server) watchFiles() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			utils.Debug(fmt.Sprintf("event: %s", event))
			if event.Op&fsnotify.Create != 0 {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					utils.Debug(fmt.Sprintf("watching new directory: %s", event.Name))
					_ = s.watcher.Add(event.Name)
					continue
				}
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 &&
				(strings.HasSuffix(event.Name, ".go") ||
					strings.HasSuffix(event.Name, ".rtml") ||
					strings.HasSuffix(event.Name, ".md") ||
					plugins.NeedsRebuild(event.Name)) {
				rebuilds.Add(1)
				utils.Info("Changes detected, rebuilding...")
				if err := build.Build(); err != nil {
					utils.Fatal("Failed to rebuild project: ", err)
				}
				if s.buildType == "ssc" {
					s.stopHost()
					if err := s.startHost(); err != nil {
						utils.Fatal("Failed to restart host server: ", err)
					}
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
	s.hostCmd = exec.Command("./host/host")
	s.hostCmd.Stdout = os.Stdout
	s.hostCmd.Stderr = os.Stderr
	return s.hostCmd.Start()
}

func (s *Server) stopHost() {
	if s.hostCmd != nil && s.hostCmd.Process != nil {
		_ = s.hostCmd.Process.Kill()
		_, _ = s.hostCmd.Process.Wait()
	}
	s.hostCmd = nil
}
