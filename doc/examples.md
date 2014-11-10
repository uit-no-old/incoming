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

We also configure the Incoming!! JavaScript library and tell it where to find the Incoming!! server. This is sadly necessary because in JavaScript code running in a browser, there is currently no way to find out from which host a JavaScript file was loaded. So if the host of the example web app and the host of the Incoming!! server are different (which might very well be, even if you use reverse proxies and load balancers and whathaveyou), the Incoming!! JavaScript library doesn't know where the Incoming!! server is unless we specifically tell it. Therefore this annoying line is necessary.

Then, the upload function. It is called from further down, when the user selects a file:

```javascript
<input type="file" id="input_file" onchange="upload_file('{{ upload_id }}', this.files[0])"/>
```

The template substitution of 'upload\_id' renders the ticket upload id directly into the HTML code, so when the onchange() callback is triggered, upload\_file() is called with the upload id the backend acquired from the Incoming!! server.

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

The second example app expands on example 1 by using more of Incoming!!'s features (pause / resume / cancel, secret backend cookie, deferred upload finish notification to Incoming!! server), and by doing more detailed inspection of the Uploader object during upload. It also demonstrates how upload tickets can be acquired dynamically from the frontend, hinting at concurrent uploads and how dynamic web apps could use Incoming!!.

The second example web app is in the [examples/2-dynamic](examples/2-dynamic) directory. Again, there are two files: [backend.py](examples/2-dynamic/backend.py) and [frontend\_tmpl.html](examples/2-dynamic/frontend_tmpl.html).


### More detailed inspection of an Uploader object

This is a very simple expansion of example 1's progress reporting, although it introduces quite a bit of code in the frontend. There are a bunch of HTML elements now for the output of numerous properties of the Uploader object. They are all updated in the 'update' callback, which is now much longer:

```javascript
    // uploader callback for updating all the HTML things
    var update = function(uploader) {
        output_state_msg.innerHTML = uploader.state_msg;
        output_progress_bar.value = uploader.frac_complete;
        output_error_msg.innerHTML = uploader.error_msg;
        output_cancel_msg.innerHTML = uploader.cancel_msg;
        output_connected.innerHTML = uploader.connected.toString();
        output_cancelling.innerHTML = uploader.cancelling.toString();
        output_cancelled.innerHTML = uploader.cancelled.toString();
        output_finished.innerHTML = uploader.finished.toString();
        output_chunks_tx.innerHTML = uploader.chunks_tx_now.toString();
        output_chunks_acked.innerHTML = uploader.chunks_acked_now.toString();
        output_chunks_ahead.innerHTML = uploader.chunks_ahead.toString();
        output_kb_tx.innerHTML = Math.round(uploader.bytes_tx / 1024);
        output_kb_acked.innerHTML = Math.round(uploader.bytes_acked / 1024);
        output_kb_ahead.innerHTML = Math.round((uploader.bytes_tx - uploader.bytes_acked) / 1024);

        input_file_select.disabled = !(uploader.cancelled || uploader.finished);
        input_pause.disabled = !uploader.can_pause;
        input_pause.checked = uploader.paused;
        input_btn_cancel.disabled = !uploader.can_cancel;
    };
```

The uploader object contains a bunch of properties you can use, including textual state and error messages (state\_msg, error\_msg, cancel\_msg), boolean states and flags (connected, can\_pause, paused, can\_cancel, cancelling, cancelled, finished), and numerical properties (frac\_complete, chunks / bytes transferred, chunks / bytes acknowledged, chunks / bytes "ahead"). The difference between "transferred" and "acknowledged" is that transferred chunks have beend sent to the server (i.e., send() has been called) but they might still reside in some buffer or be on their way, while acknowledged chunks have been acknowledged by the Incoming!! server with an "ack" message, so they have arrived at the Incoming!! server. The "\*\_ahead" properties indicate how many chunks / bytes have been sent, but not yet acknowledged. frac\_complete takes only acknowledged bytes into account, bytes that have been sent but not yet acknowledged don't count as "completed".

'chunks\_tx' and 'chunks\_acked' are only counted for the current connection, i.e., when the connection is lost and re-established, the count of transferred chunks starts again at 0. This is to avoid confusion if the server decides to change the chunk size between connections. Also, the server doesn't count the number of transferred chunks, so if the connection is lost because the browser window is closed, there is no way for the uploader to know on a reconnect how many chunks have been uploaded so far. For progress reporting, the number of transferred / acked bytes should be preferred.

Depending on the state of the upload, the callback enables or disables HTML inputs for file selection, pause, and cancel.

In addition to Uploader.onprogress and Uploader.onfinish, there are two more callbacks you can set: oncancelled and onerror. Note that onprogress is always called by Incoming!! no matter which observable property has changed, so you won't miss a cancellation or finish if all you define is an onprogress handler.


### Upload pause / resume / cancel

The frontend in example 2 lets the user pause / resume and cancel an upload. For this, there are additional HTML elements right beneath the file selector, and the 'progress update' callback (see above) dynamically enables or disables the controls based on whether pause or cancel are possible. There are also event handlers for clicks on the controls:

```javascript
    // click handler for cancel button
    input_btn_cancel.onclick = function cancel_clicked() {
        uploader.cancel("user cancelled manually");
    };

    // click handler for pause checkbox
    input_pause.onclick = function pause_clicked() {
        if (input_pause.checked) {
            uploader.pause("pause");
        } else {
            uploader.pause("unpause");
        }
    };
```

Uploader.cancel() takes one argument, a cancellation message. Uploader.pause() takes one argument which is either "pause", "unpause", or "toggle".


### Dynamic upload ticket acquisition

Instead of acquiring an upload ticket on access to the app page, example 2 can acquire upload tickets dynamically using JavaScript HTTP requests. As a consequence, upload ticked ids are no longer "hardcoded" in the HTML page, and therefore our upload\_file function no longer accepts an upload id, but only a File object - the file to upload.

```javascript
function upload_file(f) {
```

upload\_file now can't just initialize an Uploader object because it doesn't have an upload ticket yet. Instead, upload\_file does an XMLHttpRequest to the app backend, requesting an upload ticket id, and does the Uploader init/start when that request returns. In vanilla JavaScript, that looks like this:

```javascript
    // get an upload id from "my" backend (not incoming!! directly).
    // When we got the id, we start uploading.
    //
    // HTTP requests are not pretty in vanilla JavaScript, but we do it here to
    // avoid using any particular JS framework.
    var upload_id = "";
    var xhr = new XMLHttpRequest();
    xhr.open('get', "/api/frontend/request_upload?filename=" + f.name);
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4) {
            if (xhr.status == 200) {
                upload_id = xhr.responseText;

                // when we got our id, we can start uploading
                uploader = incoming.Uploader(upload_id, f);
                uploader.onprogress = update;
                uploader.onfinished = finished;
                uploader.oncancelled = update; // could do something better here
                uploader.onerror = update; // could do something better here
                uploader.start();

            } else {
                alert(xhr.responseText);
            }
        }
    };
    xhr.send(null);
```

In an actual web app, you likely use some JavaScriptframework that lets you do HTTP requests much nicer than that. In any case, the mechanism is this: request upload ticket from backend, and when that arrives, initialize and start an Uploader.

You could easily run several concurrent uploads this way (one Uploader object per upload), but to keep it simple, our example web app frontend only allows the user to upload several files sequentially.

The web app backend supports dynamic upload ticket acquisition by providing an HTTP route for the XMLHttpRequest that the frontend issues:

```python
@get('/api/frontend/request_upload')
def request_upload() :
```

In it, the backend does most of what it did before in the handler for '/', which is now much shorter. It requests an upload ticket from Incoming!!, and the answer to that request contains the upload ticket ID, which is returned to the frontend.


### "Secret backend cookie"

To keep not the middle man but at least the vandals out (the middle man can in a production setup be dealt with using SSL), example 2's backend sends a "secret" to the Incoming!! server when it requests an upload ticket:

```python
    secret = str(uuid.uuid4())

    # get an upload ticket from Incoming!!
    req_params = { "destType" : "file",
            "signalFinishURL" : "http://%s/api/backend/upload_finished" % _config["internal_app_host"],
            "removeFileWhenFinished" : "false", # we do this ourselves, by moving the file
            "signalFinishSecret" : secret,
            }
    req = requests.post("http://%s/incoming/backend/new_upload" % _config["internal_incoming_host"], params=req_params)

```

This 'secret' is later given back by the Incoming!! server when it notifies example 2's backend that the uploaded file has arrived. You can use it to verify that the request was not a bogus request from some vandal, but that it actually came from Incoming!!:

```python
@post('/api/backend/upload_finished')
def retrieve_incoming_file() :
    upload = _uploads.get(request.params["id"], None)

    if request.params["secret"] != upload["secret"] :
        print "upload_finished: wrong secret for upload id %s" % request.params["id"]
        return abort(418, "I shit you not: I am a teapot")
```

This is a quite unlikely attack vector, so we are not sure whether we keep this feature in.


### Deferred "got it" notification to Incoming!! server

When the Incoming!! server notified example 1 that the upload had arrived, example 1 would move the file and then answer that request with "done". This is okay if it doesn't take much time to move the file (max a few seconds). If processing the file takes any longer, the web app backend should instead answer the request with "wait", and then after processing is done, access Incoming!!'s POST /incoming/backend/finish\_upload. This is what example 2 does (even though it doesn't do or delegate any processing on the file).

Instead of moving the file in the 'upload\_finished' request handler, example 2 starts a new thread which does just that, and answers "wait":

```python
        incoming_path = request.params["filename"]
        dest_path = os.path.join("uploads", upload["filename"])
        answer_thread = threading.Thread(target=move_deferred,
                args=(request.params["id"], incoming_path, dest_path, 10))
        answer_thread.start()
        return "wait"
```

The 'move\_deferred' function then runs in an own thread. It moves the file, then waits until a given time has passed, and finally sends the deferred 'got it' to the Incoming!! server.

```python
    ts_start = time.time()
    shutil.move(source_path, dest_path)
    sleep_for = delay_min_s - (time.time() - ts_start)
    if sleep_for > 0 :
        time.sleep(sleep_for)

    # now tell the Incoming!! server that we are done
    req_params = { "id" : upload_id }
    req = requests.post("http://%s/incoming/backend/finish_upload" % _config["internal_incoming_host"],
        params = req_params)
```

The Incoming!! server will deem the upload successfully handed over as soon as it gets the request to finish\_upload. It will then also notify the frontend that the upload is finished.

Note that the way example 2 does it is bad design. You should not use start Python threads in Bottle requests unless your app doesn't need to scale. We just did it this way here to demonstrate the mechanism.
