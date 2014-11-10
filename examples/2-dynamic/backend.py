import sys
import time
import threading
import os
import socket
import uuid
import shutil
import requests
import click
from bottle import get, post, request, run, template, static_file, abort

_hostname = socket.getfqdn()
_config = {}
_uploads = {} # { id (str) : { "secret" : str, "filename" : str }}

@get('/')
def main_page() :
    scheme = request.urlparts[0] # 'http' or 'https'
    return template("frontend_tmpl.html",
            scheme=scheme,
            public_incoming_host=_config["public_incoming_host"],
            uploads=os.listdir("uploads"))


@get('/api/frontend/request_upload')
def request_upload() :
    filename = os.path.split(request.params["filename"])[1]
    secret = str(uuid.uuid4())

    # get an upload ticket from Incoming!!
    req_params = { "destType" : "file",
            "signalFinishURL" : "http://%s/api/backend/upload_finished" % _config["internal_app_host"],
            "removeFileWhenFinished" : "false", # we do this ourselves, by moving the file
            "signalFinishSecret" : secret,
            }
    req = requests.post("http://%s/incoming/backend/new_upload" % _config["internal_incoming_host"], params=req_params)

    # if status code is OK, the request returns the upload id in the return body. If the status
    # code is an error code, the body contains an error message.
    if req.status_code != requests.codes.ok :
        return abort(500, "incoming!! error: %d %s" % (req.status_code, req.text))
    upload_id = req.text

    _uploads[upload_id] = { "secret" : secret, "filename" : filename }

    return upload_id


@post('/api/backend/upload_finished')
def retrieve_incoming_file() :
    # if you have a webserver / reverse proxy in front of your web app, you
    # might want to make it block external access to URLs starting with
    # /backend
    # In any case, we can check the secret we gave to Incoming!! earlier.
    upload = _uploads.get(request.params["id"], None)
    if upload == None :
        return abort(404, "There's no upload with that ID")

    if request.params["secret"] != upload["secret"] :
        print "upload_finished: wrong secret for upload id %s" % request.params["id"]
        return abort(418, "I shit you not: I am a teapot")

    # If upload was successful and not cancelled, start a thread that
    # moves the uploaded file to its destination path, waits a bit, and then
    # signals Incoming!! that the file is completely handed over. Notifying
    # Incoming!! deferred like this (as opposed to example 1) is fine to
    # take some time.
    ret = ""
    if request.params["cancelled"] != "yes" :
        incoming_path = request.params["filename"]
        dest_path = os.path.join("uploads", upload["filename"])
        answer_thread = threading.Thread(target=move_deferred,
                args=(request.params["id"], incoming_path, dest_path, 10))
        answer_thread.start()
        return "wait"
    else :
        # we don't care. request.params["cancelReason"] contains a text describing
        # why the upload was cancelled. It also doesn't matter what we answer.
        print "Upload %s was cancelled: %s" % (request.params["id"],
            request.params["cancelReason"])
        del _uploads[request.params["id"]]
        return ""


def move_deferred(upload_id, source_path, dest_path, delay_min_s) :
    # move file, then perhaps wait until delay_min_s seconds have passed since
    # invocation
    ts_start = time.time()
    shutil.move(source_path, dest_path)
    sleep_for = delay_min_s - (time.time() - ts_start)
    if sleep_for > 0 :
        time.sleep(sleep_for)

    # now tell the Incoming!! server that we are done
    req_params = { "id" : upload_id }
    req = requests.post("http://%s/incoming/backend/finish_upload" % _config["internal_incoming_host"],
        params = req_params)

    # we get either "ok" or an error message as an answer
    if req.text != "ok" :
        print "Incoming!! error on finish (doesn't matter): %d %s" % (req.status_code, req.text)

    # Finally, remove upload from our hash table. This probably isn't thread
    # safe! Good that this is only an example ;)
    del _uploads[upload_id]


@get('/uploads/<filename:path>')
def send_uploaded_file(filename) :
    return static_file(filename, root="uploads")


@click.command()
@click.option('--public_incoming_host', help='(public) incoming host name[:port].',
        default=_hostname+':4000')
@click.option('--internal_incoming_host', help='(internal) incoming host name[:port].',
        default='localhost:4000')
@click.option('--public_app_host', help='(public) app host name[:port].',
        default=_hostname+':4002')
@click.option('--internal_app_host', help='(internal) app host name[:port] visible to incoming.',
        default='localhost:4002')
@click.option('--port', help='port web app should listen on',
        default=4002)
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
