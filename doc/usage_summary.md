This is orphaned text from the 'main page' which was replaced by the usage example section.


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

