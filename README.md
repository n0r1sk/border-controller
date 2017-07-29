[![Build Status](https://travis-ci.org/n0r1sk/border-controller.svg?branch=edge)](https://travis-ci.org/n0r1sk/border-controller) [![Anchore Image Overview](https://anchore.io/service/badges/image/68f00f08cde7b43f90ad3ce9a3a48bf282e649e1bd6854df47e7875f9d1f5882)](https://anchore.io/image/dockerhub/n0r1skcom%2Fborder-controller%3A1.0.1)

# border-controller
This is a Nginx based ingress border controller with automatic configuration reload based on Docker swarm service discovery for on-premise, but not limited to, use.

# Why?
The current problem in the Docker swarm infrastructure is, that the swarm mesh network does currently not support sticky connections.

# Configuration

```
debug: true
general:
  check_intervall: 10
  resources:
    testcontexta:
      context: /context/a
      port: 8080
      task_dns: tasks.testa.app
    testcontextb:
      context: /context/b
      port: 9090
      task_dns: tasks.testb.app
```

## debug
If parameter is not provided, default value set to false.

## check_intervall
If parameter is not provided, default value set to 30 secons. Time duration is in seconds.

# version
1.0
