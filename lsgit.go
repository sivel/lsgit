package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Handler struct {
	Root string
}

func (h *Handler) LsGit(w http.ResponseWriter, req *http.Request) {
	pathParts := strings.Split(req.URL.Path, "/")

	if len(pathParts) != 3 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	repoPath := fmt.Sprintf("%s.git", filepath.Join(h.Root, pathParts[1], pathParts[2]))

	_, err := os.Stat(repoPath)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	var stdout bytes.Buffer
	cmd := exec.Command("git", "ls-tree", "-r", "HEAD")
	cmd.Dir = repoPath
	cmd.Stdout = &stdout

	cmd.Run()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	var response []string
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue
		}
		response = append(response, parts[1])
	}

	b, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var port string
	var root string
	flag.StringVar(&port, "port", ":3001", "HOST:PORT to listen on, HOST not required to listen on all addresses")
	flag.StringVar(&root, "root", "/tmp", "Root of where all users live")
	flag.Parse()

	h := Handler{
		Root: root,
	}

	http.HandleFunc("/", h.LsGit)

	log.Fatal(http.ListenAndServe(port, nil))
}
