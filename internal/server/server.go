package server

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/rfwlab/rfw/internal/build"
	"github.com/rfwlab/rfw/internal/utils"
)

type Server struct {
	Port    string
	Host    bool
	stopCh  chan os.Signal
	watcher *fsnotify.Watcher
}

func NewServer(port string, host bool) *Server {
	return &Server{
		Port:   port,
		Host:   host,
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
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 &&
				(strings.HasSuffix(event.Name, ".go") || strings.HasSuffix(event.Name, ".rtml")) {
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
