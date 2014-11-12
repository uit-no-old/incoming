Incoming!! APIs
===============

There are three places in the system where there are APIs: the JavaScript library exposes a JavaScript API to your frontend, the Incoming!! server exposes an HTTP API to your backend, and your backend must expose one HTTP function to the Incoming!! server.


Incoming!! JavaScript library (frontend)
----------------------------------------

Your web app's frontend loads the JavaScript library like any other JavaScript library, for example by using a `<script>` tag:

```html
<script src="http[s]://INCOMING_HOSTNAME/incoming/frontend/incoming.js"></script>
```

After loading the JavaScript file, there is one more object in the global namespace: `incoming`. It contains the library.


### The `incoming` library object

The `incoming` library exposes two functions. Other than that, there is nothing of interest.


### `incoming.set_server_hostname( hostname )`

Sets the hostname where to find the Incoming!! server. This should be set to the same value as INCOMING\_HOSTNAME you used when loading the JavaScript file. It is sad that we need this function at all, but in JavaScript running in a browser, there is no neat way to find out from which host a JavaScript file has been loaded from.

You need to call this function only once, but it has to be called before any `Uploader` objects are created. We recommend calling it as soon as the page has loaded, for example in `window.onload`. 


### incoming.Uploader



Incoming!! server HTTP API (backend)
------------------------------------

The POST request goes to the Incoming!! server (whose host:port we know from the command line) and specifies a two parameters:

* 'destType': destination type - file or some other sort of storage object. At present, only 'file' is supported, but we will probably add support for storage systems such as Ceph in the future.
* 'signalFinishURL': URL the Incoming!! server should POST to when the file has arrived.
* 'removeFileWhenFinished': should the Incoming!! server, when all is done, remove the file or not? In this example, the backend moves the file in the filesystem, so the Incoming!! server should not (try to) remove the uploaded file.
* 'backendSecret': you can protect the 'signalFinishURL' from bogus accesses a bit by specifying a string here which the Incoming!! server will send back when accessing the URL later. This won't stop the man in the middle, but if you run Incoming!! server and web app on one host or if you use SSL between the two, this is useful.

requests.post() is synchronous and returns when the request has been answered. So the next thing to do is to assert that we got a good answer, and to get our upload ticket from it:
