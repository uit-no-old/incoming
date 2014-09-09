/*
Incoming!!

Roadmap:
- listening IP and port should be configurable
- file uploads go into a temporary directory that can be specified in config
- js lib: callbacks settable as object fields (like in many js APIs), not in constructor.
- implement clean up
- implement cancel, error handling, resume
- initialization of upload module: clean temp dir (in case of app crash and restart)
- make code agnostic to which CWD it is called from
- document. make sure to mention that adblock (at least on chrome) causes high
  CPU usage. recommend to disable adblock for the page that uses incoming.
--> 0.1 finished

- (optional) file verification after upload: checksum in browser and backend, then
  assert that checksum is the same. Most likely error scenario: user updated file on
  disk while upload was running.
- go through ways for web app to retrieve file
  - available on filesystem? just give it the path to the file then (good enough if
    incoming!! serves only one web app). web app must move file away (or copy it), then
    tell incoming!! that it is finished. This is what we have now.
  - web app could download the file (very bad idea with most web apps, as it takes time,
    but this should be easy to implement)
  - if stored in cloud storage (ceph?): object id will work. coolest solution, as the file
    is stored in the right place right away

open questions:
- web app frontend must know name of incoming!! server. how will it get to know that?
    - for now: web app backend knows. html includes URL to js file.
- Incoming!! js code must know name of incoming!! server. how will it know?
    - for now, there is a function set_server_hostname in the incoming lib that must
	  be called by the web app frontend. Can we simplify this?
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"source.uit.no/lars.tiede/incoming/upload"
)

type appVarsT struct {
	uploaders upload.UploaderPool
	config    *appConfigT
}

var appVars *appVarsT

/* NewUploadHandler receives an http request from a webapp wanting to do
an upload, and makes an Uploader for it. It responds with the uploader's id
(string).
*/
func NewUploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got new upload request")
	// read upload parameters from request
	destType := r.FormValue("destType")
	if destType == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "destType not given")
		return
	}
	signalFinishURL, err := url.ParseRequestURI(r.FormValue("signalFinishURL"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "signalFinishURL invalid: %s", err.Error())
		return
	}
	removeFileWhenFinished, err := strconv.ParseBool(
		r.FormValue("removeFileWhenFinished"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "removeFileWhenFinished invalid: %s", err.Error())
		return
	}
	signalFinishSecret := r.FormValue("signalFinishSecret")
	// signalFinishSecret is optional, so it's fine when it's empty

	// make (and pool) new uploader
	storageDirAbsolute, _ := filepath.Abs(appVars.config.StorageDir)
	uploader := upload.NewUploadToLocalFile(appVars.uploaders,
		storageDirAbsolute, signalFinishURL,
		removeFileWhenFinished, signalFinishSecret)

	// answer request with id of new uploader
	fmt.Fprint(w, uploader.GetId())
	return
}

func ServeJSFileHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./incoming_jslib.js") // TODO make this robust. need dir
	// from somewhere
}

func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Note that this can be called by both backend and frontend

	// fetch uploader for given id
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "id not given")
		return
	}
	uploader, ok := appVars.uploaders.Get(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "id unknown")
		return
	}

	// let uploader cancel (async because this method should return quickly)
	tellWebAppBackend := true
	if strings.Contains(r.URL.String(), "backend/") {
		tellWebAppBackend = false
	}
	go uploader.Cancel(tellWebAppBackend, "Cancelled by request",
		time.Duration(appVars.config.HandoverTimeoutS)*time.Second)
	fmt.Fprint(w, "ok")
	// when uploader is done cancelling, it will send "upload finished" to web
	// app backend if necessary, so we are done here
	return
}

func main() {
	log.SetFlags(log.Lshortfile)

	// --- init application-wide things (config, data structures)
	appVars = new(appVarsT)
	var err error

	// load config
	appVars.config, err = LoadConfig("./config.yaml") // TODO: LoadConfig should figure out from where to get the config file (do it like ansible does)
	if err != nil {
		log.Printf("Exit program")
		return
	}

	// init uploader pool
	appVars.uploaders = upload.NewLockedUploaderPool()

	// --- set up http server
	routes := mux.NewRouter()
	routes.HandleFunc("/backend/new_upload", NewUploadHandler).
		Methods("POST")
	routes.HandleFunc("/backend/cancel_upload", CancelUploadHandler).
		Methods("POST")
	routes.HandleFunc("/frontend/cancel_upload", CancelUploadHandler).
		Methods("POST")
	//routes.HandleFunc("/backend/finish_upload", FinishUploadHandler).
	//		Methods("POST")
	// TODO: write handler
	routes.HandleFunc("/frontend/upload_ws", websocketHandler).
		Methods("GET")
	routes.HandleFunc("/frontend/incoming.js", ServeJSFileHandler).
		Methods("GET")
	// TODO: write handler (check ServeContent or ServeFile)

	// --- run server forever
	serverHost := fmt.Sprintf("0.0.0.0:%d", appVars.config.IncomingPort)
	log.Printf("Will start server on %s", serverHost)
	log.Fatal(http.ListenAndServe(serverHost, routes))
}
