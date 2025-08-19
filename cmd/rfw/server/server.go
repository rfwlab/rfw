package server

import (
	"bufio"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/rfwlab/rfw/cmd/rfw/build"
	"github.com/rfwlab/rfw/cmd/rfw/plugins"
	"github.com/rfwlab/rfw/cmd/rfw/utils"
)

var rebuilds = expvar.NewInt("rebuilds")

type Server struct {
	Port    string
	Host    bool
	Debug   bool
	stopCh  chan os.Signal
	watcher *fsnotify.Watcher
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

	fs := http.FileServer(http.Dir("."))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		utils.LogServeRequest(r)
		s.handleFileRequest(w, r, fs)
	})

	if s.Debug {
		http.Handle("/debug/vars", expvar.Handler())
		http.HandleFunc("/debug/pprof/", pprof.Index)
		http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		http.HandleFunc("/debug/pprof/profile", pprof.Profile)
		http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		http.HandleFunc("/debug/pprof/trace", pprof.Trace)
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
	utils.PrintStartupInfo(s.Port, localIP, s.Host)

	go func() {
		if err := http.ListenAndServe(":"+s.Port, nil); err != nil {
			utils.Fatal("Server failed: ", err)
		}
	}()

	go s.listenForCommands()

	<-s.stopCh
	utils.Info("Server stopped.")
	return nil
}

func (s *Server) handleFileRequest(w http.ResponseWriter, r *http.Request, fs http.Handler) {
	if _, err := os.Stat("." + r.URL.Path); os.IsNotExist(err) {
		http.ServeFile(w, r, "./index.html")
	} else {
		fs.ServeHTTP(w, r)
	}
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
			utils.PrintStartupInfo(s.Port, localIP, s.Host)
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
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			utils.Info(fmt.Sprintf("Watcher error: %v", err))
		case <-s.stopCh:
			s.watcher.Close()
			return
		}
	}
}
