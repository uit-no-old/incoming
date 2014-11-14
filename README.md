Incoming!!
==========

Large file uploads with web browsers are frustrating because you can not implement them in a straightforward and painless way. When implemented wrong, they stall your web app and might even break it, and implementing them right is not easily done. Incoming!! handles large file uploads for your web apps so you don't have to. In the browser, it chops up a large file into little pieces, transfers those over to its own server application, which puts them together, stores them, and finally hands the uploaded file over to your web app backend. Disconnects during upload are no problem, and explicit pause / resume is also supported. With Incoming!!, both the complexity and performance impact of large file uploads are off your web app's back.

Incoming!! consists of a server application and a JavaScript client library. The server can run alongside your web app backend or centrally in your organization, and the JavaScript client is used directly in your web app's frontend in the browser.

When you want to upload a large file, your web app backend first fetches an upload ticket from the Incoming!! server. Then, using that ticket, your frontend can use the Incoming!! JavaScript library to send the large file to the Incoming!! server. When the upload is finished, the Incoming!! server hands the file over to your web app backend.

Your web app backend and the Incoming!! server communicate through simple HTTP requests: Incoming!! provides an endpoint for handing out upload tickets, and issues a request to your web app backend when an upload is finished. To its JavaScript library, Incoming!! exposes a WebSocket server. Incoming!! also hosts the JavaScript library file.

The Incoming!! server is implemented in Go, and the JavaScript library in, well, JavaScript, without use of external libraries. The whole thing is free and open source, licensed under the permissive TODO LICENSE.


Status
------

Incoming!! is in development. Most of it is in place and works. It can already be used, but we don't consider it and the API stable yet.

At the present stage of development, Incoming!! can already be deployed together with an individual application, or centrally in your organization. However, for the latter case to be viable, Incoming!! should be able to scale out with several server instances in order to provide ample upload bandwidth. This has always been in the backs of our heads during design and implementation, and is the next major feature we will implement.


Usage example
-------------

Simple backend code in Python (using the tiny [Bottle](http://bottlepy.org/) web framework) can look like the following.

Request an upload ticket from the Incoming!! server like this (using the [Requests](http://python-requests.org) library):

```python
req = requests.post("http://INCOMING_HOSTNAME/incoming/backend/new_upload",
                     params = { "signalFinishURL" : "http://APP_HOSTNAME/api/backend/upload_finished" })
```

In the request for a ticket, you tell the Incoming!! server which URL to POST to later when the upload is finished.

The request returns the upload ticket ID, which you somehow give to your frontend. For example, if you are just answering a request for a file upload page, you could simply render the upload ticket ID into a template.

```python
upload_id = req.text
return template("upload_page_template.html", upload_id = upload_id)
```

Later, when the upload is finished, the Incoming!! server will POST to the URL you specified above. In that request, you get the path of the uploaded file so you can move the file to its destination. This is how a handler for that URL can look like in your web app backend:

```python
@post('/api/backend/upload_finished')
def retrieve_incoming_file() :
    upload_id = request.params["id"]

    if request.params["cancelled"] != "yes" :
        incoming_path = request.params["filename"]
        shutil.move(incoming_path, os.path.join("uploads", request.params["filenameFromBrowser"]))
    else :
        # we don't care. request.params["cancelReason"] contains a text describing
        # why the upload cancelled. It also doesn't matter what we answer.

    return "done"
```

After you return "done", the Incoming!! server notifies your frontend that the upload is all done. Then both your backend and your frontend know that the upload is finished.

Speaking of your frontend, here's what you need to do there. First, you have to load the Incoming!! JavaScript library:

```html
<script src="http[s]://INCOMING_HOSTNAME/incoming/frontend/incoming.js"></script>
```

Then, you need some sort of file input, for example a file input field. To that, you can attach an event handler to kick off an upload as soon as the user chooses a file:

```html
<input type="file" id="input_file" onchange="upload_file('{{ upload_id }}', this.files[0])"/>
```

Here, we have rendered in the upload ticket id in the backend. You could of course obtain a ticket in other ways, for example with an extra bit of JavaScript that does an HTTP request to your backend (one of our example apps does that).

You can then configure and start an upload like this:

```javascript
function upload_file(upload_id, f) {
    // before we do any uploads, we have to tell the incoming!! js library the
    // host:port of the incoming!! server
    incoming.set_server_hostname("INCOMING_HOSTNAME");

    // define a callback for when upload is finished (i.e., the web app backend
    // got the file)
    var finished = function(uploader) {
        alert("yay, upload is finished");
    };

    // initialize and start uploader
    var uploader = incoming.Uploader(upload_id, f);
    uploader.onfinished = finished;
    uploader.start();
}
```

When `uploader.start()` is called, Incoming!! will do its thing. When everything is done, that is, when the file has been uploaded and handed over to your web app backend, your "upload is finished" callback is called. Then you know that your app has gotten the file.

This is basically it. We ship two example web apps that serve as more comprehensive code examples, handling errors and covering more of Incoming!!'s features. In most usage scenarios there is also the webserver / reverse proxy config to take care of. For that, we also document and ship an example.


Documentation
-------------

* [System overview: motivation, design, integration eight-miles-up](doc/overview.md)
* [Installation: manual or automated installation of Incoming!!, example apps, and an example reverse proxy](doc/installation.md)
* [Getting started: example web apps using Incoming!!](doc/examples.md)
* [Incoming!! Frontend and Backend API](doc/api.md)
* [Important notes for developers and users](doc/notes.md)


Changelog, roadmap etc.
-----------------------

* [Changelog](doc/changelog.md)
* [Roadmap](doc/roadmap.md)


License
-------

Incoming!! is licensed under the TODO FIND A DAMN LICENSE
