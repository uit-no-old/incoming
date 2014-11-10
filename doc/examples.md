Example web apps using Incoming!!
=================================

We provide two example web apps together with the Incoming!! source code. Example 1 is kept as simple as possible. Example 2 showcases dynamic acquisition of upload tickets, hinting at use in dynamic web apps and the possibility of concurrent uploads. Example 2 also uses most of Incoming!!'s features such as upload pause/resume and more detailed inspection of the upload progress.


## Example web app 1: simple file upload page rendered by backend

In this example web app, the user gets a file upload page when navigating to the page ('/'). She can select a file using a standard HTML file selector, and when she has done that, the file is uploaded using Incoming!!. When the web app backend is notified by the Incoming!! server that the upload has arrived, it moves the file to another location from where the user can download it later.

Let's have a look at how all of this is implemented. The example web app is in the [examples/1-simple](examples/1-simple) directory. There are two files: [backend.py](examples/1-simple/backend.py) and [frontend\_tmpl.html](examples/1-simple/frontend_tmpl.html). The backend is written in Python and uses the tiny and dead simple [Bottle](http://bottlepy.org/) web framework. You don't need to know Bottle in order to understand what is happening - if you have worked with a web framework before, you can probably understand very well what's going on. The backend also uses the [Requests](http://python-requests.org) and [Click](http://click.pocoo.org/) libraries - Requests is for doing HTTP requests and very simple (you will understand it instantly), and Click is mostly for command line parsing (not necessary to understand here). The frontend template is a simple HTML file upload page with some JavaScript in order to use Incoming!!.


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

The POST request goes to the Incoming!! server (whose host:port we know from the command line) and specifies two parameters:

* 'signalFinishURL': URL the Incoming!! server should POST to when the file has arrived.
* 'removeFileWhenFinished': should the Incoming!! server, when all is done, remove the file or not? In this example, the backend moves the file in the filesystem, so the Incoming!! server should not (try to) remove the uploaded file. This is optional, and the default value is "true".

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

The frontend is the upload page that is displayed in the user's browser. It is served by the backend when '/' is accessed. The backend renders [frontend\_tmpl.html](examples/1-simple/frontend_tmpl.html), a template written with Bottle's template syntax, which is, unsurprisingly, very similar to most template syntaxes. Let's have a look what Incoming!!-related things example app 1 is doing in that template.

First, the Incoming!! JavaScript library is loaded:

```html
<script src="{{ scheme }}://{{ public_incoming_host }}/incoming/frontend/incoming.js"></script>
```

'scheme' and 'public\_incoming\_host' are template substitutions reflecting the system setup and make sure that the browser finds the library.

Next comes a script block in which we define a few functions. First, we make sure that the Incoming!! JavaScript library could actually be loaded. If that is the case, we have a global variable called 'incoming':

```javascript
window.onload = function() {
    // make sure that the incoming lib has loaded. This might not be the case
    // if some browser add-on has blocked it, or if incoming.js could not be
    // downloaded from the Incoming!! server for some reason.
    if (typeof incoming === 'undefined') {
        var output_state_msg = document.getElementById("stats_state_msg");
        output_state_msg.innerHTML = "Did not load incoming lib - was it blocked? Is the incoming server offline?";
        return;
    }

    // before we do any uploads, we have to tell the incoming!! js library the
    // host:port of the incoming!! server
    incoming.set_server_hostname("{{ public_incoming_host }}");
};
```

We also configre the Incoming!! JavaScript library and tell it where to find the Incoming!! server. This is sadly necessary because in JavaScript code running in a browser, there is currently no way to find out from which host a JavaScript file was loaded. So if the host of the example web app and the host of the Incoming!! server are different (which might very well be, even if you use reverse proxies and load balancers and whathaveyou), the Incoming!! JavaScript library doesn't know where the server is unless we specifically tell it. Therefore this annoying line is necessary.

Then, the upload function. It is called from further down, when the user selects a file:

```javascript
<input type="file" id="input_file" onchange="upload_file('{{ upload_id }}', this.files[0])"/>
```

The template substitution of 'upload\_id' will render the ticket upload id directly into the HTML code, so when the onchange() callback is triggered, upload\_file() is called with the upload id the backend acquired from the Incoming!! server.

The upload\_file function sets up the file upload and starts it. It receives the (textual) upload id and a [File](https://developer.mozilla.org/en/docs/Web/API/File) object as parameters:

```javascript
function upload_file(upload_id, f) {
```

It first checks whether a file was actually passed to the function (a browser might call this funtion without passing in a File). Then, it defines two callback functions: one for all sorts of observable upload progress (we use this for a progress bar), and one for when the upload is finished (we display a message):

```javascript
    // define a callback for all sorts of progress in the uploader
    var update = function(uploader) {
        document.getElementById("stats_state_msg").innerHTML = uploader.state_msg;
        document.getElementById("stats_progress_bar").value = uploader.frac_complete;
    };

    // define a callback for when upload is finished (i.e., the web app backend
    // got the file)
    var finished = function(uploader) {
        var output_node = document.getElementById("output_finished");
        output_node.innerHTML = "<p><b>Upload is finished. Reload this page to upload another file or to see the uploaded file in the list below.</b></p>";
        output_node.hidden = false;
    };
```

With these functionalities in place, we can make an Uploader object, set it up with our callback functions, and start uploading:

```javascript
    // initialize uploader
    var uploader = incoming.Uploader(upload_id, f);
    uploader.onprogress = update;
    uploader.onfinished = finished;

    // when everything is set up, unleash uploader. It will do its thing in possibly
    // many asynchronous steps, and call the callbacks when appropriate.
    uploader.start();
```

Since this simple example app can only upload one file (because it got only one upload ticket from the backend rendered directly into the HTML file), we forbid further clicks on the file selector directly after having started the file upload:

```javascript
    // once we're uploading, the user may not select another file (at least in
    // this simple example).
    document.getElementById("input_file").disabled = true;
```

To upload another file, the user will have to reload the page after the upload is complete. Then, the backend acquires a new ticket.

That's basically it. The rest of the template is what little HTML we need: file selector, progress output, and a template-generated list of uploaded files that can be downloaded.


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

