Example web apps using Incoming!!
=================================

We provide two example web apps together with the Incoming!! source code. Example 1 is kept as simple as possible. Example 2 showcases dynamic acquisition of upload tickets, hinting at use in dynamic web apps and the possibility of concurrent uploads. Example 2 also uses most of Incoming!!'s features such as upload pause/resume and more detailed inspection of the upload progress.


## Example web app 1: simple file upload page rendered by backend

In this example web app, the user gets a file upload page when navigating to the page ('/'). She can select a file using a standard HTML file selector, and when she has done that, the file is uploaded using Incoming!!. When the web app backend is notified by the Incoming!! server that the upload has arrived, it moves the file to another location from where the user can download it later.

Let's have a look at how all of this is implemented. The example web app is in the [examples/1-simple](../examples/1-simple) directory. There are two files: [backend.py](../examples/1-simple/backend.py) and [frontend\_tmpl.html](../examples/1-simple/frontend_tmpl.html). The backend is written in Python and uses the tiny and dead simple [Bottle](http://bottlepy.org/) web framework. You don't need to know Bottle in order to understand what is happening - if you have worked with a web framework before, you can probably understand very well what's going on. The backend also uses the [Requests](http://python-requests.org) and [Click](http://click.pocoo.org/) libraries - Requests is for doing HTTP requests and very simple (you will understand it instantly), and Click is mostly for command line parsing (pretty irrelevant here). The frontend template is a simple HTML file upload page with some JavaScript in order to use Incoming!!.


### Backend

We first take a closer look at the backend, starting with the function that makes the main page (the one the user gets when navigating to '/'). The main page is supposed to be the application's frontend - an HTML file with the file upload page and some JavaScript to use Incoming!!. Before delivering the page, the backend must acquire an upload ticket from the Incoming!! server. The upload ticket ID is then rendered into the page that is sent to the user's browser. That way, the frontend does not need to request a ticket ID for the file upload at any point because it got it delivered along with the file upload page.

So the first thing the web app backend does when handling the HTTP request is to request an upload ticket from Incoming!!:

```python
@get('/')
def main_page() :
    # get an upload ticket from Incoming!!
    req_params = {
            "signalFinishURL" : "http://%s/api/backend/upload_finished" % 
                _config["internal_app_host"],
            "removeFileWhenFinished" : "false" # we do this ourselves, by moving the file
            }
    req = requests.post("http://%s/incoming/backend/new_upload" % _config["internal_incoming_host"], params=req_params)
```

The POST request goes to the Incoming!! server (whose host:port we know from the command line) and specifies a two parameters:

* 'signalFinishURL': URL the Incoming!! server should POST to when the file has arrived.
* 'removeFileWhenFinished': should the Incoming!! server, when all is done, remove the file or not? In this example, the backend moves the file in the filesystem, so the Incoming!! server should not (try to) remove the uploaded file.

requests.post() is synchronous, so it returns when the request has been answered. The next thing to do is to assert that we got a good answer, and to get our upload ticket from it:

```python
    # if status code is OK, the request returns the upload id in the return
    # body. If the status code is an error code, the body contains an error
    # message.
    if req.status_code != requests.codes.ok :
        return abort(500, "incoming!! error: %d %s" % (req.status_code, req.text))
    upload_id = req.text
    _uploads[upload_id] = True # all we need is the key
```

We store the upload ticket id in a global hash table to keep track of it.

If we didn't bail out with an error page because something went wrong, we can now render the page template and send it to the user's browser:

```python
    scheme = request.urlparts[0] # 'http' or 'https'
    return template("frontend_tmpl.html",
            scheme=scheme,
            public_incoming_host=_config["public_incoming_host"],
            upload_id = upload_id,
            uploads=os.listdir("uploads"))
```

Note that the upload ticket ID is rendered into the template that is delivered to the user's browser.

Besides acquiring upload tickets, the app backend needs to provide a URL that the Incoming!! server can POST to when the upload has arrived. This URL has been sent to Incoming!! when requesting a ticket. Parameters of the request include the ticket ID of the upload that has arrived, whether the upload has been cancelled or not, and what filename the uploaded file had in the user's browser. The app backend can use all this information to first assert that an upload with the reported ticket ID actually exists (always mind the vandals...):

```python
@post('/api/backend/upload_finished')
def retrieve_incoming_file() :
    # if you have a webserver / reverse proxy in front of your web app, you
    # might want to make it block external access to URLs starting with
    # /backend
    upload = _uploads.get(request.params["id"], None)
    if upload == None :
        return abort(404, "There's no upload with that ID")
```

Then, in case the upload wasn't cancelled, the backend can move the file to wherever it wants (here, an 'uploads' directory):

```python
    if request.params["cancelled"] != "yes" :
        incoming_path = request.params["filename"]
        shutil.move(incoming_path, os.path.join("uploads", request.params["filenameFromBrowser"]))
        ret = "done"
    else :
        # we don't care. request.params["cancelReason"] contains a text describing
        # why the upload cancelled. It also doesn't matter what we answer.
        print "Upload %s was cancelled: %s" % (request.params["id"],
            request.params["cancelReason"])
        ret = ""
```

And finally, the backend returns "done" in case of success, or anything (it doesn't matter) if the upload was cancelled.

The rest of the backend code for example 1 is a static file download handler so that the uploaded files can be downloaded again, and command line and server startup stuff.


### Frontend


## Example 2: dynamic ticket acquisition, and using most of Incoming!!'s features

TODO that snippet below is from the old example 1, before simplifying it even more

```python
@get('/')
def main_page() :
    # get an upload ticket from Incoming!!
    secret = str(uuid.uuid4())
    req_params = { "destType" : "file",
            "signalFinishURL" : "http://%s/api/backend/upload_finished" % _config["internal_app_host"],
            "removeFileWhenFinished" : "false", # we do this ourselves, by moving the file
            "signalFinishSecret" : secret,
            }
    req = requests.post("http://%s/incoming/backend/new_upload" % _config["internal_incoming_host"], params=req_params)
```

The POST request goes to the Incoming!! server (whose host:port we know from the command line) and specifies a few parameters:

* 'destType': destination type - file or some other sort of storage object. At present, only 'file' is supported, but we will probably add support for storage systems such as Ceph in the future.
* 'signalFinishURL': URL the Incoming!! server should POST to when the file has arrived.
* 'removeFileWhenFinished': should the Incoming!! server, when all is done, remove the file or not? In this example, the backend moves the file in the filesystem, so the Incoming!! server should not (try to) remove the uploaded file.
* 'signalFinishSecret': you can protect the 'signalFinishURL' from bogus accesses a bit by specifying a string here which the Incoming!! server will send back when accessing the URL later. This won't stop the man in the middle, but if you run Incoming!! server and web app on one host or if you use SSL between the two, this is useful.

requests.post() is synchronous and returns when the request has been answered. So the next thing to do is to assert that we got a good answer, and to get our upload ticket from it:

```python
    # if status code is OK, the request returns the upload id in the return
    # body. If the status code is an error code, the body contains an error
    # message.
    if req.status_code != requests.codes.ok :
        return abort(500, "incoming!! error: %d %s" % (req.status_code, req.text))
    upload_id = req.text
    _uploads[upload_id] = { "secret" : secret }
```


TODO move this to ex 2.
If the backend needs some time to process the file before the upload can be considered finished, for example if the file has to be moved to a different filesystem (we don't recommend this at all, but as a matter of fact our automated example setup will do just that), the request can also be answered with "wait", and then Incoming!!'s `POST /incoming/backend/finish_upload` be accessed later.

