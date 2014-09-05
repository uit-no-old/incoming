Usage of Incoming!! in your web app
===================================

In the sequence of what things happen during a large file upload, this is how you implement large file uploads using Incoming!!:

1) your frontend (the browser part) tells your backend that it wants to upload a file (in whichever way you want - how you implement this bit is entirely up to you)
2) your backend requests a ticket for an incoming upload from Incoming!! using an HTTP request (GET /backend/new_upload). In that request, your backend tells Incoming!! where it should store the file, and which URL to use to signal the upload's completion to your backend (more on that in step 5). Incoming!!s answer to that request contains a ticket ID.
3) your backend tells your frontend the ticket ID for the file upload (how you implement this is again up to you - the straightforward way to do this would be to do steps 1 and 3 with one HTTP request and answer done in JavaScript).
4) your frontend calls a JavaScript function from the js file that comes with Incoming!!, which will do the magic of chopping up the file and sending the parts individually to Incoming!!. Hooks are available for your frontend to track upload progress and completion (more on that in step 7).
5) Incoming!! sends an HTTP request to your backend to signal the upload's completion. Your backend can now do whatever it wants with that file.
6) optional: if your backend configured the uploaded file to be deleted by Incoming!! (you can do that in step 2), it must call Incoming!!s HTTP DELETE /backend/delete_upload?ticket_id=xxx function
7) in your frontend, the upload function you called in step 4 calls a hook your frontend supplies that signals the upload's completion. Now both your banckend and your frontend know that the upload is done.

TODO: this is actually a special case - the dynamic case. It is possible to do it easier, when the web app backend gets an upload id *before* anything is sent to the browser. Maybe describe that (first).
