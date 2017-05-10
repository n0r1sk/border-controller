[![Build Status](https://travis-ci.org/n0r1sk/ingress-controller.svg?branch=master)](https://travis-ci.org/n0r1sk/ingress-controller)

# border-controller
This is a Nginx based ingress border controller with automatic configuration reload based on Docker swarm service discovery for on-premise, but not limited to, use.

# Why?
The current problem in the Docker swarm infrastructure is, that the swarm mesh network does currently not support sticky connections.

