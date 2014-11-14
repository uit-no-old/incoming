/*
Incoming!!

Roadmap:
- document. make sure to mention that adblock (at least on chrome) causes high
  CPU usage. recommend to disable adblock for the page that uses incoming.
- make versioned URLs
- choose a license
--> 0.1 finished
- improve logging
- error codes

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
	"path"
	"path/filepath"
	"strconv"
	"time"

	"bitbucket.org/kardianos/osext"
	"github.com/gorilla/mux"

	"source.uit.no/star-apt/incoming/upload"
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

	// upload to file or... (nothing else supported yet)
	destType := r.FormValue("destType") // 'file' or nothing. Default: file
	if destType == "" {
		destType = "file"
	}

	// which URL to POST to when file is here
	signalFinishURL, err := url.ParseRequestURI(r.FormValue("signalFinishURL"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "signalFinishURL invalid: %s", err.Error())
		return
	}

	// should we remove the file when it's all over or not?
	removeFileWhenFinishedStr := r.FormValue("removeFileWhenFinished")
	if removeFileWhenFinishedStr == "" { // true or false. Default: true
		removeFileWhenFinishedStr = "true"
	}
	removeFileWhenFinished, err := strconv.ParseBool(removeFileWhenFinishedStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "removeFileWhenFinished invalid: %s", err.Error())
		return
	}

	// secret cookie to POST to finish URL later
	backendSecret := r.FormValue("backendSecret") // optional, "" if not given

	// make (and pool) new uploader
	storageDirAbsolute, _ := filepath.Abs(appVars.config.StorageDir)
	uploader := upload.NewUploadToLocalFile(appVars.uploaders,
		storageDirAbsolute, signalFinishURL,
		removeFileWhenFinished, backendSecret,
		time.Duration(appVars.config.UploadMaxIdleDurationS)*time.Second)

	// answer request with id of new uploader
	fmt.Fprint(w, uploader.GetId())
	return
}

func ServeJSFileHandler(w http.ResponseWriter, r *http.Request) {
	programDir, _ := osext.ExecutableFolder()
	filePath := path.Join(programDir, "incoming_jslib.js")
	http.ServeFile(w, r, filePath)
}

func FinishUploadHandler(w http.ResponseWriter, r *http.Request) {
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

	// assert that 'backend secret string' matches (if it's not given, it's an
	// empty string, which might be just fine)
	if uploader.GetBackendSecret() != r.FormValue("backendSecret") {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "backendSecret not given or wrong")
		return
	}

	// tell uploader that handover is done
	err := uploader.HandoverDone()

	// return error message or "ok"
	if err != nil {
		fmt.Fprint(w, err.Error())
	} else {
		fmt.Fprint(w, "ok")
	}
}

func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
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

	// assert that 'backend secret string' matches (if it's not given, it's an
	// empty string, which might be just fine)
	if uploader.GetBackendSecret() != r.FormValue("backendSecret") {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "backendSecret not given or wrong")
		return
	}

	// let uploader cancel
	err := uploader.Cancel(false, "Cancelled by request",
		time.Duration(appVars.config.HandoverTimeoutS)*time.Second)

	// on success, clean up and return "ok". On failure, return error message
	if err == nil {
		uploader.CleanUp()
		fmt.Fprint(w, "ok")
	} else {
		w.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprintf(w, "%v", err)
	}

	return
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// --- init application-wide things (config, data structures)
	appVars = new(appVarsT)
	var err error

	// load config
	appVars.config, err = LoadConfig()
	if err != nil {
		log.Printf("Couldn't load config!")
		log.Fatal(err)
		return
	}

	// init upload module
	err = upload.InitModule(appVars.config.StorageDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	// init uploader pool
	appVars.uploaders = upload.NewLockedUploaderPool()

	// --- set up http server
	routes := mux.NewRouter()
	routes.HandleFunc("/incoming/backend/new_upload", NewUploadHandler).
		Methods("POST")
	routes.HandleFunc("/incoming/backend/cancel_upload", CancelUploadHandler).
		Methods("POST")
	routes.HandleFunc("/incoming/backend/finish_upload", FinishUploadHandler).
		Methods("POST")
	routes.HandleFunc("/incoming/frontend/upload_ws", websocketHandler).
		Methods("GET")
	routes.HandleFunc("/incoming/frontend/incoming.js", ServeJSFileHandler).
		Methods("GET")

	// --- run server forever
	serverHost := fmt.Sprintf("%s:%d", appVars.config.IncomingIP,
		appVars.config.IncomingPort)
	log.Printf("Will start server on %s", serverHost)
	log.Fatal(http.ListenAndServe(serverHost, routes))
}
