Roadmap
=======


Version 0.2
-----------

* make Incoming!! scalable: support several backends behind a (redirecting) load balancer. For that, we probably need:
  * reference load balancer setup. keep it simple, use a generic one. Round-robin redirection in nginx should be perfectly fine as a start.
  * networked uuid pool (just use redis or something like that?) to map which Incoming!! instance 'has' a certain upload
  * method(s) to 'move' uploader objects between Incoming!! instances
    * "pull": "I have to continue doing this upload. Give me the uploader object"
    * would be nice to be able to 'empty' an instance of all uploads to be able to do zero downtime (rolling) server updates. But where would they go? Should we be able to "push" uploader objects to arbitrary Incoming!! instances, and then later, when the client reconnects, let the instance that was chosen by the (rather dumb) load balancer "pull" the uploader object from wherever it was temporarily stored?

Any or all of the following (cans that can probably be kicked down the road):

* improved logging
* improved acquisition of app configuration (check TODO note in source)
* proper error codes everywhere
* incoming.set\_server\_hostname default could be better. Let Incoming!! backend figure this out an use template substitution in the JS library?
* HTTP API: return values in response bodies are okay, but what do with error messages? More or less fitting HTTP error code and error message in response body is not too nice. Use JSON in responses?


Version 0.3
-----------

* support for cloud storage that does not expose itself as a filesystem. Ceph? That amazon protocol "everybody" is compatible with?


Wishlist (far future?)
----------------------

* (optional) file verification after uploads that happened in several sessions: checksum in browser and backend, assert that they are identical. Most likely error scenario: user updated file on his device between upload sessions.


Back to [main page](../README.md)
