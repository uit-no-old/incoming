Incoming!! APIs
===============

There are three places in the system where there are APIs: the JavaScript library exposes a JavaScript API to your frontend, the Incoming!! server exposes an HTTP API to your backend, and your backend must expose one HTTP function to the Incoming!! server.


Incoming!! JavaScript library (frontend)
----------------------------------------

Your web app's frontend loads the JavaScript library like any other JavaScript library, for example by using a `<script>` tag:

```html
<script src="http[s]://INCOMING_HOSTNAME/incoming/0.1/frontend/incoming.js"></script>
```

After loading the JavaScript file, there is one more object in the global namespace: `incoming`. It contains the library.


### The `incoming` library object

The `incoming` library only exposes two functions.


#### `incoming.set_server_hostname( hostname )`

Sets the hostname where to find the Incoming!! server. You need to call this function if the Incoming!! server name is not the same as your web app's server name. In that case, this should be set to the same value as INCOMING\_HOSTNAME you used when loading the JavaScript file. If this function is not called, Incoming!! will use `window.location.host` as its server.

You need to call this function only once, but it has to be called before any `Uploader` objects are created. We recommend calling it as soon as the page has loaded, for example in `window.onload`. 


#### `incoming.Uploader( upload_id, file )`

Creates and returns an uploader object, which will do all the magic for one file. `upload_id` is an Incoming!! upload ticket ID that you somehow got from your backend (see [system overview](overview.md), [examples](examples.md)). `file` is a [File](https://developer.mozilla.org/en/docs/Web/API/File) object that you can get from an HTML file selector or file drop area.

If you want to upload several files concurrently, use several uploader objects, one for each file. Each file needs its own upload ticket.


### `Uploader` objects

An uploader object uploads one file. Each uploader object needs its own upload ticket. Several uploader objects can upload several files concurrently.

In an uploader object, there are many properties and flags that you can use for inspection. In order to track progress, you can set one callback that is called whenever any observable property changes its value (`onprogress`). There are three other settable callbacks that are called in addition to `onprogress` on important events: `onfinished`, `oncancelled`, and `onerror`. To control an upload, three functions are available: `start`, `pause`, and `cancel`.

Typically, you would create an uploader object, define one or more callbacks, and then call start().


#### Properties

All properties should be treated as read-only.

* `filename` - name of the file, without any path information
* `bytes_total` - length of the file, in bytes
* `bytes_tx` - number of bytes that have been sent to the Incoming!! server
* `bytes_acked` - number of bytes that we know have arrived at the Incoming!! server
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
* `paused` - true if upload is currently paused
* `can_pause` - true if upload can currently be paused. Upload can only be paused when chunks are being transferred to the Incoming!! server. It is not possible to pause uploads when they are already being handed over to the web app backend.
* `cancelling` - true if the upload is currently cancelling. This is the case when the uploader has sent a cancellation message to the Incoming!! server and is waiting for a reply.
* `cancelled` - true if the upload has been cancelled.
* `can_cancel` - true if the upload can currently be cancelled. Uploads that are currently being handed over and paused uploads can not be cancelled.


#### Callbacks

In an uploader object, you can set any of the following callback functions. All functions accept one parameter: the upload object.

* `onprogress` - called whenever one or several of the observable properties and flags change.
* `onfinished` - called (in addition to onprogress) when the 'finished' flag value is changed from false to true.
* `oncancelled` - called (in addition to onprogress) when the 'cancelled' flag value is changed from false to true.
* `onerror` - called (in addition to onprogress) when an error occurs.


#### Functions

* `start()` - starts the upload.
* `pause( what )` - pauses, unpauses, or toggles pause. 'what' can either be 'pause', 'unpause', or 'toggle'.
* `cancel( reason )` - cancels the upload. 'reason' is a string and should explain why the caller cancels the upload.


Incoming!! server HTTP API (backend)
------------------------------------

The Incoming!! server exposes a few HTTP functions to your web app backend, for acquiring upload tickets, cancelling uploads, and deferred finish notifications.

On success, all functions return a 200 status code, and the body of the response contains the function's return value. On failure, the functions return some 4xx or 5xx code, and the body of the response contains an error message.

In order to secure the interaction between your web app backend and Incoming!!, the backend API offers you to use an optional session 'backend secret' (the upload ticket ID is not considered secret). This 'secret' - just an arbitrary string you can specify in your web app backend - is passed around on all communication between your web app backend and the Incoming!! server. It helps to rule out bogus accesses to the various HTTP functions, but it can't do anything against the middle man. To keep that one at bay, you need to encrypt communication between your web app backend and Incoming, for example with SSL. In combination, secure end-to-end communication and our shared session secret should sufficiently secure all interaction between your web app backend and Incoming!! when the network between the two can't be trusted.


### Functions

#### `POST /incoming/0.1/backend/new_upload`

Acquire an upload ticket. Parameters (passed as form values):

* `signalFinishURL` - URL the Incoming!! server should POST to when the file has arrived. For details on that function you have to provide check the 'Your web app backend HTTP API' section below.
* `destType` (optional, defaults to 'file') - destination type. 'file' or some other sort of storage object. At present, only 'file' is supported, but we will probably add support for storage systems such as Ceph in the future.
* `removeFileWhenFinished` (optional, defaults to 'true') - should the Incoming!! server, when all is done, remove the uploaded file or not? If your web app backend moves the file to another location during handover, you should set this to 'false'.
* `backendSecret` (optional, defaults to '') - an arbitrary string that will henceforth be used as the backend secret for this upload.

Return value (passed as response body): upload ticket id - a UUID string.


#### `POST /incoming/0.1/backend/cancel_upload`

Cancel an ongoing upload if it is not already too late for that (i.e., if Incoming!! is not already handing the file over to your web app backend). Parameters (passed as form values):

* `id` - upload ticket id of the upload.
* `backendSecret` (optional, defaults to ''): - shared secret string for this upload

Return value (passed as response body): 'ok'


#### `POST /incoming/0.1/backend/finish_upload`

URL to POST to when deferred finish notification is used, i.e., when Incoming!!'s request to your web app backend's signalFinishURL is answered with 'wait' and not 'done'.

* `id` - upload ticket id of the upload.
* `backendSecret` (optional, defaults to ''): - shared secret string for this upload

Return value (passed as response body): 'ok'


Your web app backend HTTP API
-----------------------------

Your web app backend must expose one HTTP function to the Incoming!! server: the one the Incoming!! server calls to notify your web app backend about an uploaded file that has arrived (called 'signalFinishURL' above). The name of it doesn't matter (let's just call it `/api/backend/hand_over_upload` here) because you tell the name to the Incoming!! server.


#### `POST /api/backend/hand_over_upload`

Accessed by the Incoming!! server when an upload has arrived and can be handed over, or when either the user or Incoming!! has cancelled the upload. Answer this request either with 'wait' or 'done'. When the answer is 'wait', your web app backend must POST to Incoming!!'s /incoming/0.1/backend/finish\_upload later in order to signal to Incoming!! that the upload is finished. When the answer is "done", Incoming!! considers the upload finished as soon as it gets that response.

Parameters (passed as form values):

* `id` - upload ticket id of the upload.
* `backendSecret` - shared secret string for this upload (defaults to '' if there was no shared secret for this upload).
* `filename` - path to the uploaded file (as Incoming!! sees it).
* `filenameFromBrowser` - name of the file that the client uploaded
* `cancelled` - whether the upload was cancelled or not. "yes" or "no"
* `cancelReason` - informal text describing why upload was cancelled

Return value (passed as response body): 'wait' or 'done'.


Back to [main page](../README.md)
