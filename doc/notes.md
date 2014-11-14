Important notes for developers and users
========================================

There are a few things very worth noting that don't fit anywhere else in the documentation.

CPU usage in the browser
------------------------

Incoming!! uses significant CPU time, might even bog down a user's device if one of the following criteria is met:

* SSL is used, and bandwidth is high: encrypting data uses CPU, encrypting lots of data uses lots of CPU. To aggravate the problem, at the time of this writing SSL encryption on Google Chrome on the author's mac is single-threaded and thus becomes a bottleneck quickly. The author's uploads maxed out at ~30MB/s with more bandwidth available and, interestingly, the (weak ass one core VM) server's CPU not being maxed out.
* browser plugins such as AdBlock are used: the author doesn't know which plugins exactly cause high CPU usage during uploads, but AdBlock certainly does. On the author's mac, the CPU maxes out and no more than 5MB/s is achieved.
* developer tools in the browser are running: similar situation as with AdBlock (or other related plugins)
