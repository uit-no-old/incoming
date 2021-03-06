import sys
import os
import socket
import uuid
import shutil
import requests
import click
from bottle import get, post, request, run, template, static_file, abort

_hostname = socket.getfqdn()
_config = {}
_uploads = {} # { id (str) }

@get('/')
def main_page() :
    # get an upload ticket from Incoming!!
    req_params = {
            "signalFinishURL" : "http://%s/api/backend/upload_finished" % 
                _config["internal_app_host"],
            "removeFileWhenFinished" : "false" # we do this ourselves, by moving the file
            }
    req = requests.post("http://%s/incoming/0.1/backend/new_upload" % _config["internal_incoming_host"],
            params=req_params)

    # if status code is OK, the request returns the upload id in the return
    # body. If the status code is an error code, the body contains an error
    # message.
    if req.status_code != requests.codes.ok :
        return abort(500, "incoming!! error: %d %s" % (req.status_code, req.text))
    upload_id = req.text
    _uploads[upload_id] = True # all we need is the key

    scheme = request.urlparts[0] # 'http' or 'https'
    return template("frontend_tmpl.html",
            scheme=scheme,
            public_incoming_host=_config["public_incoming_host"],
            upload_id = upload_id,
            uploads=os.listdir("uploads"))

@post('/api/backend/upload_finished')
def retrieve_incoming_file() :
    # if you have a webserver / reverse proxy in front of your web app, you
    # might want to make it block external access to URLs starting with
    # /backend
    upload = _uploads.get(request.params["id"], None)
    if upload == None :
        return abort(404, "There's no upload with that ID")

    # If upload was successful and not cancelled, move uploaded file to
    # destination path. Note that you need access to both paths, and that this
    # operation should be quick (i.e., source and destination paths should be
    # on the same volume).
    ret = ""
    if request.params["cancelled"] != "yes" :
        incoming_path = request.params["filename"]
        shutil.move(incoming_path, os.path.join("uploads", request.params["filenameFromBrowser"]))
        ret = "done"
    else :
        # we don't care. request.params["cancelReason"] contains a text describing
        # why the upload was cancelled. It also doesn't matter what we answer.
        print "Upload %s was cancelled: %s" % (request.params["id"],
            request.params["cancelReason"])
        ret = ""

    del _uploads[request.params["id"]]
    return ret # we can return "done" or "wait" here. If "done", then for incoming
               # the upload is history now. If "wait", then incoming will wait until
               # we access POST /incoming/backend/finish_upload
   # What happens now (in case of success): the incoming!! backend will signal
   # the frontend that the upload is done. Through a JS callback, your frontend
   # will be able to know as well.


@get('/uploads/<filename:path>')
def send_uploaded_file(filename) :
    return static_file(filename, root="uploads")


@click.command()
@click.option('--public_incoming_host', help='(public) incoming host name[:port].',
        default=_hostname+':4000')
@click.option('--internal_incoming_host', help='(internal) incoming host name[:port].',
        default='localhost:4000')
@click.option('--public_app_host', help='(public) app host name[:port].',
        default=_hostname+':4001')
@click.option('--internal_app_host', help='(internal) app host name[:port] visible to incoming.',
        default='localhost:4001')
@click.option('--port', help='port web app should listen on',
        default=4001)
def run_server(public_incoming_host, internal_incoming_host,
        public_app_host, internal_app_host, port) :
    global _config
    _config["public_incoming_host"] = public_incoming_host
    _config["internal_incoming_host"] = internal_incoming_host
    _config["public_app_host"] = public_app_host
    _config["internal_app_host"] = internal_app_host
    _config["port"] = port

    if not os.path.isdir("uploads") :
        os.mkdir("uploads")

    run(host='0.0.0.0', port=_config["port"])

if __name__ == "__main__" :
    run_server()
