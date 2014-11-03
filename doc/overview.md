System overview
===============

Motivation
----------

Large file uploads should be trivial, but sadly aren't. Browsers and web apps / web servers alike still struggle with this. This is because HTTP has no built-in support for streamed uploads, or even only binary file uploads. Instead, files are uploaded as part of regular HTTP requests. Most webservers and web frameworks do not stream HTTP requests, but process them whole. This causes many problems with large file uploads on the server side. For example, a webserver often caches the whole file that comes as part of an HTTP request before handing it, again whole, to the web app backend. This can take a lot of time and bogs down the web app backend when it receives and processes the file. Another problem is that a large file upload either succeeds or fails as a whole: if anything happens during the time the request is sent from the browser to the server, the whole file upload has failed and must be repeated as a whole. Yet another problem is that the application running in the browser usually doesn't know anything about the progress of an ongoing upload. All your web app frontend knows is that a request (for example a form submit) is going on that takes a very long time.

One way of avoiding most of these problems is to chop up a large file into small pieces and to upload these pieces individually. This is not standardized in any way, and there is no built-in support for this functionality in any popular browsers or webservers yet. Therefore, custom code is needed on both client and server sides. On the client side, some JavaScript must chop up the file and send the parts individually to the server. On the server side, some web server must receive the parts and assemble them into one file again. This has been implemented many times for different web frameworks and applications. However, web app backends that handle large file uploads are still bogged down by them. Instead of being able to answer many small requests, they are busy handling only very few large requests (file uploads). Web applications that handle large file uploads must therefore scale out early and wide. 

Incoming!! factors out large file uploads into an own application. Any web application can use Incoming!!, which makes its functionality available to the web application frontend via a JavaScript file, and to the web application backend via an HTTP API. It is possible to run Incoming!! alongside an individual web app, or centrally for your whole organization that might run many web apps written in various languages and frameworks. This way, both the complexity and the performance impact of large file uploads can be taken away from your web app backends.


Basic design
------------

Incoming!! factors out large file uploads from your web app. It does this by running its own server application which communicates with your web app backend through simple HTTP requests, and by providing a JavaScript library that you use directly in your frontend in the browser to upload files.

Here is an overview of the components of a typical web app that uses Incoming!!, with Incoming!! components being the slightly colored ones, and the rest being *your* components that you would also have without Incoming!!.

![components][fig-components]

As in most web applications, the design is split into two distinct sides: the browser side (stuff running in your users' web browsers) and the server side (everything running on your servers or at your cloud provider).

On the browser side we find your web app frontend, which is typically some sort of HTML5 application with logic implemented in JavaScript that is downloaded to your users' browsers. This web app frontend can access Incoming!!'s functionality through the Incoming!! JavaScript library, which is a JavaScript file that your web app frontend loads with a `<script>` HTML tag.

On the server side, we have your web app backend, that is some application handling HTTP requests coming in from your users' browsers. Since you deal with large files, you also have some sort of storage, for example a spacious file system or a clustered storage system such as Ceph. Incoming!!'s server lives alongside your app's backend and the storage. These three components can communicate with each other over the network.

The browser and server side communicate with each other over the Internet using HTTP and WebSockets. It is likely that you funnel all that traffic (your web app's traffic *and* Incoming!! traffic) through one or several reverse proxies. This is not strictly necessary, though, and for the sake of simplicity we omit the reverse proxy from the following discussions. The basic Incoming!! design doesn't change significantly with reverse proxies, it is just simpler to introduce the system without them. The example web apps we ship with the source code use a reverse proxy.

The following picture shows the main data flows in the system, with arrow widths roughly indicating the amount of data we typically expect to flow through the system.

PIC 2

Incoming!! does the heavy lifting, while your web app frontend and backend communicate with Incoming!! through thin APIs that don't exchange much data. Most importantly, your web app backend is relieved from handling large amounts of data at any point. It never needs to touch uploaded files directly. If your web app backend and the Incoming!! server run on different machines, your web app backend is also shielded from file upload traffic. Note though that the Incoming!! server and your web app backend both need access to your storage, which in the distributed case has to be accessible over the network. Typical storage systems for this would be simple filesystem network shares like CIFS or NFS (on a third machine) or a clustered storage system such as Ceph.

By the way, if your web app is supposed to serve large files to your users, it is likewise a good idea to not let your web app backend handle any large amounts of data (if that is possible). Typically, you would use a webserver that has a sendfile (or similarly named) extension to let the webserver do the heavy lifting of delivering large files to clients.


A large file upload with Incoming!!
-----------------------------------

Suppose your web app backend wants to let a client upload a large file, for example by rendering and then sending a page with a file input field, or by answering a specific AJAX request by the client. This is the sequence in which the file upload with Incoming!! happens:


SEQ 1

1. your backend requests an upload ticket from Incoming!! using an HTTP request. In that request, your backend tells Incoming!! which URL to use to signal the upload's completion back to your backend (more on that in step 5). Incoming!!'s answer to the request contains the upload ticket ID.

SEQ 2

2. your backend lets your frontend know the ticket ID for the file upload. How you implement this is up to you - you could for example render this into a page template that contains a file upload form, or implement it as an extra HTTP request if you want to allow your frontend to dynamically request upload tickets.

SEQ 3

3. your frontend sets up and starts the file upload using Incoming!!'s JavaScript library.

SEQ 4

4. The Incoming!! JavaScript library establishes a connection to the Incoming!! server and sends the file over, in many small chunks (a). The Incoming!! server assembles the file again and stores it into some data storage (currently, a file system) that is accessible from both Incoming!! server and your web app backend (b).

SEQ 5

5. Incoming!! sends an HTTP request to your backend to signal the upload's completion (a). Your backend can now do whatever it wants with that file, for example move it (b).

SEQ 6

6. optional: if your backend responds to Incoming!!'s request with "done", the upload is done and Incoming!! tells the frontend that it's done. However, your web app backend can also answer "wait" if processing the file takes some time, and then call Incoming!!'s HTTP `POST /backend/finish_upload` function later to tell Incoming!! that the frontend can be informed about the upload being finished.

SEQ 7

7. The Incoming!! server tells the frontend that the upload is handed over to the application (a). In your frontend, a callback is called (b). Now both your backend and your frontend know that the upload is done.

[fig-components]: figures/components.svg
