package main

import (
	"flag"
	"log"
	"net/http"
)

type Handler struct {
	reposRootPath string
	internalApi   InternalApi
	logger        Logger
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := &HttpProtocolHandler{w, req, h.reposRootPath}

	responseText, err := execute(handler, h.internalApi, h.logger)
	if err != nil {
		h.WriteError(responseText, err)
	}
}

func main() {
	var (
		reposRootPath  = flag.String("r", ".", "Directory containing git repositories")
		internalApiUrl = flag.String("u", "http://localhost:3000/api/internal", "...")
		addr           = flag.String("l", ":80", "Address/port to listen on")
	)
	flag.Parse()

	internalApi := NewGitoriousInternalApi(*internalApiUrl)
	logger := nil
	http.Handle("/", &Handler{*reposRootPath, internalApi, logger}) // TODO: just use handlerFunc
	log.Fatal(http.ListenAndServe(*addr, nil))
}
