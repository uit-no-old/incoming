Incoming!!
==========

Large file upload, usable from any web app.


Motivation
----------

Large file uploads aren't trivial, browsers and webapps / webservers alike still struggle with it. This is because HTTP has no built-in support for streamed uploads, or even only binary file uploads. Instead, files are usually uploaded as part of a regular HTTP request. Webservers and most webapps do not stream HTTP requests, but process them whole. This causes obvious problems with big files.

One way of avoiding the problem is to chop up the file into small pieces and upload these pieces individually using many HTTP requests or a websocket connection. This is not standardized in any way, and there is no built-in support for this functionality on any popular browsers or webservers yet. Therefore, custom code is needed on the client side (that is, JavaScript that chops up the file and sends the individual parts) and on the server side (a web server / app that receives the parts and assembles them into a large file).

In order to not have to reinvent the wheel for each application that uses large file uploads, Incoming!! factors out large file uploads into an own application / web server. Any web application can use Incoming!!, which makes its functionality available to the web application frontend via a JavaScript file, and to the web application backend via an HTTP API. It is possible to run Incoming!! alongside an individual web app, or centrally for a whole organization that runs many web apps.

There is already a large file upload solution out there that UiT uses ([filesender](https://www.filesender.org/)), but it is intended to be used as a standalone application and is not easily integratable into other webapps. Instead of gutting filesender and seeing whether we can factor out the bare upload functionality and then add the bits we need (communication with the web app that uses filesender, javascript code for the web app's frontend), we decided that it is easier to just roll our own.


Running the examples
--------------------

Several eamples in the examples/ directory are provided that demonstrate how you can use Incoming!! in your web app. To get these up and running, you need python, pip, and a few dependencies. This is how you get them installed in a virtual environment:

    $ cd examples
    examples$ mkdir py-env
    examples$ virtualenv py-env
    examples$ source py-env/bin/activate
    (py-env) examples$ pip install -r pip-req.txt

Now you can run the example backends in whichever shell you have sourced py-env/bin/activate in.


Integration into your web app
-----------------------------

See [doc/usage.md](doc/usage.md).


Deployment centrally or as part of your web app
-----------------------------------------------

See doc/deployment.md

TODO. Describe how to compile and install (maybe have binary available?). Describe how to set up nginx for accepting frontend URLs from anywhere, but backend URLs only internally. nginx setup different when used centrally or alongside app (i.e., can webserver config be tighter integrated with web app when incoming comes alongside it?)
