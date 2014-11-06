Incoming!!
==========

Large file uploads with web browsers are frustrating because you can not implement them in a straightforward and painless way. When implemented wrong, they stall your web app and might even break it, and implementing them right is not easily done. Incoming!! handles large file uploads for your web apps so you don't have to. In the browser, it chops up a large file into little pieces, transfers those over to its own server application, which puts them together, stores them, and finally hands the uploaded file over to your web app backend. Disconnects during upload are no problem, and explicit pause / resume is also supported. With Incoming!!, both the complexity and performance impact of large file uploads are off your web app's back.

Incoming!! consists of a server application and a JavaScript client library. The server can run alongside your web app backend or centrally in your organization, and the JavaScript client is used directly in your web app's frontend in the browser. When you want to upload a large file, your web app backend first fetches an upload ticket from the Incoming!! server. Then, using that ticket, your frontend can use the Incoming!! JavaScript library to send the large file to the Incoming!! server. When the upload is finished, the Incoming!! server hands the file over to your web app backend.

Your web app backend and the Incoming!! server communicate through simple HTTP requests: Incoming!! provides an endpoint for handing out upload tickets, and issues a request to your web app backend when an upload is finished. To its JavaScript library, Incoming!! exposes a WebSocket server. Incoming!! also hosts the JavaScript library file.

The Incoming!! server is implemented in Go, and the JavaScript library in, well, JavaScript, without use of external libraries. The whole thing is free and open source, licensed under the permissive TODO LICENSE.

At the present stage of development, Incoming!! can already be deployed together with an individual application, or centrally in your organization. However, for the latter case to be viable, Incoming!! should be able to scale out with several server instances in order to provide enough upload bandwidth. This use case is always in the back of our heads, but not fully implemented yet.


Usage example
-------------

TODO the crucial bits from example 1, verbatim.


Usage summary
-------------

TODO when the usage example is in place, do I still want this section (here)?

Install the Incoming!! server alongside your web app, or centrally for your organization. With the Incoming!! server in place, you do the following in your web app backend to use it:

* you request an *upload ticket* from the Incoming!! server with an HTTP POST request to `/incoming/backend/new_upload`. Parameters include a URL that the Incoming!! server should access later when the upload is finished, so that your web app backend will know when the upload is done. The Incoming!! server answers with a *ticket ID*.
* the ticket ID has to be communicated to the frontend somehow. How you do this is up to you. Two typical approaches are your backend rendering the ticket ID into a template for a file upload page, or to do it dynamically with AJAX. (The reason for letting you do this in your app instead of letting Incoming!! do this is to avoid burdening Incoming!! with any sort of user authentication - this is your app's job, not Incoming!!'s.)
* you need to provide a route for the aforementioned "upload is finished" URL. It will get an HTTP POST request from the Incoming!! server when an upload is finished. Parameters include the path of the uploaded file on the filesystem that both Incoming!! server and your web app backend have access to. Now your web app backend can rename the uploaded file, update your database etc. When the HTTP request returns, the Incoming!! server assumes that the upload is finished.

You can of course customize the whole process a bit, but this is the gist of it.

In your web app's frontend (i.e., the HTML + JavaScript stuff running in the user's browser), you do the following:

* you need to load the Incoming!! JavaScript library, typically using a `<script>` tag. It is hosted on the Incoming!! server.
* when you want to upload a file, your web app frontend needs to get a ticket ID from your backend (which in turn got it from the Incoming!! server - see above). 
* when you got a ticket, you can use `incoming.Uploader()` to initialize an Uploader object, and `Uploader.start()` to start uploading. You can provide callbacks for progress, completion, etc., and a bunch of variables are available in an Uploader object for you to track and inspect uploads.
* as soon as an Uploader object calls the completion callback, you know that the file has been uploaded and handed over to your web app backend.

Again, you can customize the process a bit, but that's the gist.

When you set all of this up, you should make sure that Incoming!!'s `new_upload` URL is not accessible from the outside. This is to avoid unauthorized uploads - only your app backends, which take care of authorisation, should be able to access that URL. Similarly, you can shield your backend's "upload is finished" URL from accesses from the outside. One typical way of doing all of this is to set up your webserver or reverse proxy so that accesses "from the outside" to these URLs is forbidden. You can also set up your webserver to block *all* accesses to these URLs, and configure your web app backend and Incoming!! to communicate with each other directly. The example web apps in the source repository use this setup.


Documentation
-------------

* [System overview: motivation, design, integration eight-miles-up](doc/overview.md)
* [Installation](doc/installation.md)
* [Example web apps using Incoming!!](doc/examples.md)
* [Client and Server API](doc/api.md)
