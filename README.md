[![Build Status](https://travis-ci.org/n0r1sk/border-controller.svg?branch=edge)](https://travis-ci.org/n0r1sk/border-controller)

# border-controller
This is a Nginx based ingress border controller with automatic configuration reload based on Docker swarm service discovery for on-premise, but not limited to, use.

# Why?
The current problem in the Docker swarm infrastructure is, that the swarm mesh network does currently not support sticky connections.

# Configuration

You cannot use **ingress_service_name** and **stack_service_task_dns_name** at the same time! You have to decide if you want to use the **docker_controller** or the Docker Swarm Network Overlay DNS descovery service.

```
debug: true
general:
  swarm:
    docker_hosts:
      - your.docker.host
    docker_host_dns_domain: docker.host
    ingress_service_name: ingress_lb
    stack_service_task_dns_name: tasks.your_app
    stack_service_port: 8080
    docker_controller:
      api_key: your-docker-controller-api-key
      exposed_port: 21212
  check_intervall: 10
```

## debug
If parameter is not provided, default value set to false.

## check_intervall
If parameter is not provided, default value set to 30 secons. Time duration is in seconds.

