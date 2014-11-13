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

The `incoming` library only exposes two functions.


#### `incoming.set_server_hostname( hostname )`

Sets the hostname where to find the Incoming!! server. This should be set to the same value as INCOMING\_HOSTNAME you used when loading the JavaScript file.

You need to call this function only once, but it has to be called before any `Uploader` objects are created. We recommend calling it as soon as the page has loaded, for example in `window.onload`. 

It is sad that we need this function at all, but in JavaScript running in a browser, there is no good way to find out which host a JavaScript file has been loaded from.


#### `incoming.Uploader( upload_id, file )`

Creates and returns an uploader object, which will do all the magic for one file. `upload_id` is an Incoming!! upload ticket ID that you somehow got from your backend (see [system overview](overview.md), [examples](examples.md)). `file` is a [File](https://developer.mozilla.org/en/docs/Web/API/File) object that you can get from an HTML file selector or file drop area.

If you want to upload several files concurrently, use several uploader objects, one for each file. Each file needs its own upload ticket.


### `Uploader` objects

An uploader object uploads one file. Each uploader object needs its own upload ticket. Several uploader objects can upload file concurrently.

In an uploader object, there are many properties and flags that you can use for inspection. In order to track progress, you can set one callback that is called whenever any observable property changes its value (`onprogress`). There are three other settable callbacks that are called in addition to `onprogress` on important events: `onfinished`, `oncancelled`, and `onerror`. To control an upload, three functions are available: `start`, `pause`, and `cancel`.

Typically, you would create an uploader object, define one or more callbacks, and then call start().


#### Properties

* `filename` - name of the file, without any path information
* `bytes_total` - length of the file, in bytes
* `bytes_tx` - number of bytes that have been sent to the Incoming!! server
* `bytes_acked` - number of bytes that have arrived at the Incoming!! server
* `bytes_ahead` - number of bytes that have been sent to the Incoming!! server but have not yet arrived (they might be in some buffer outside our control on either side, or they might be on their way, or they might have arrived but the Incoming!!'s server acknowledgement is still on its way back).
* `frac_complete` - fraction of upload that has arrived at the Incoming!! server. This is a numerical value between 0 and 1. When the value is 1, the file has been uploaded to the Incoming!! server, but that doesn't mean that the upload is finished: that is only the case after Incoming!! has handed the file over to your backend. There is no measure of progress for that; handover starts when the file has arrived at Incoming!!, and it ends when your backend reports back to Incoming!! that it is finished getting the file. Depending on your application, that might take milliseconds or ages.
* `chunks_tx_now` - number of chunks (messages containing file data) that have been sent during the current connection. When the connection is lost and re-established, this count goes back to 0. Chunk sizes may vary between connections (in the future, perhaps also during connections).
* `chunks_acked_now` - number of chunks that have arrived at the Incoming!! server during the current connection. When the connection is lost and re-established, this count goes back to 0.
* `chunks_ahead` - number of chunks that have been sent but have not been acknowledged yet.
* `state_msg` - text describing the current state of the uploader.
* `cancel_msg` - if the upload is cancelled, cancel\_msg contains a text describing the reason for the cancellation
* `error_code` - if an error has occurred, error\_code contains a numerical error code. At present, there are no error codes yet :(
* `error_msg` - if an error has occurred, error\_msg contains a textual error message.


#### Flags

All flags are boolean.

* `connected` - true if uploader is connected to Incoming!! server, false if not
* `finished` - true if upload is finished, i.e., it has successfully been handed over to the web app backend
* `paused` - true is upload is currently paused
* `can_pause` - true if upload can currently be paused. Upload can only be paused when chunks are being transferred to the Incoming!! server. It is not possible to pause uploads when they are already being handed over to the web app backend.
* `cancelling` - true if the upload is currently cancelling. This is the case when the uploader has sent a cancellation message to the Incoming!! server and is waiting for a reply.
* `cancelled` - true if the upload has been cancelled.
* `can_cancel` - true if the upload can currently be cancelled. Uploads that are currently being handed over and paused uploads can not be cancelled.


#### Callbacks

In an uploader object, you can set any of the following callback functions. All functions accept one parameter: the upload object.

* `onprogress` - called whenever any of the observable properties and flags changes its value.
* `onfinished` - called (in addition to onprogress) when the 'finished' flag value is changed from false to true.
* `oncancelled` - called (in addition to onprogress) when the 'cancelled' flag value is changed from false to true.
* `onerror` - called (in addition to onprogress) when an error occurs.


#### Functions

* `start()` - starts the upload.
* `pause( what )` - pauses, unpauses, or toggles pause. 'what' can either be 'pause', 'unpause', or 'toggle'.
* `cancel( reason )` - cancels the upload. 'reason' is a string and should explain why the caller cancel the upload.




Incoming!! server HTTP API (backend)
------------------------------------

The POST request goes to the Incoming!! server (whose host:port we know from the command line) and specifies a two parameters:

* 'destType': destination type - file or some other sort of storage object. At present, only 'file' is supported, but we will probably add support for storage systems such as Ceph in the future.
* 'signalFinishURL': URL the Incoming!! server should POST to when the file has arrived.
* 'removeFileWhenFinished': should the Incoming!! server, when all is done, remove the file or not? In this example, the backend moves the file in the filesystem, so the Incoming!! server should not (try to) remove the uploaded file.
* 'backendSecret': you can protect the 'signalFinishURL' from bogus accesses a bit by specifying a string here which the Incoming!! server will send back when accessing the URL later. This won't stop the man in the middle, but if you run Incoming!! server and web app on one host or if you use SSL between the two, this is useful.

requests.post() is synchronous and returns when the request has been answered. So the next thing to do is to assert that we got a good answer, and to get our upload ticket from it:
