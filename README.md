![Build Status](https://img.shields.io/docker/pulls/n0r1skcom/border-controller.svg) ![Build Status](https://img.shields.io/docker/automated/n0r1skcom/border-controller.svg) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fn0r1sk%2Fborder-controller.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fn0r1sk%2Fborder-controller?ref=badge_shield)
 ![Build Status](https://img.shields.io/docker/build/n0r1skcom/border-controller.svg) [![Build Status](https://travis-ci.org/n0r1sk/border-controller.svg?branch=edge)](https://travis-ci.org/n0r1sk/border-controller)

[![Anchore Image Overview](https://anchore.io/service/badges/image/68f00f08cde7b43f90ad3ce9a3a48bf282e649e1bd6854df47e7875f9d1f5882)](https://anchore.io/image/dockerhub/n0r1skcom%2Fborder-controller%3A1.0.1)

# border-controller
This is a Nginx based ingress border controller with automatic configuration reload based on Docker swarm DNS service discovery for on-premise, but not limited to, use.

# Why?
The current problem in the Docker swarm infrastructure is, that the swarm mesh network does currently not support sticky connections. We know, that for example [Traefik](https://docs.traefik.io/) exists, which covers many if not all problems mentioned here. But currently, Traefik does not support TCP load balancing and you will have services, which are not using the HTTP protocol. Furthermore, there will be setups, where you won't let Traefik communicate with the Docker swarm manager.

Please don't get this wrong, but if Traefik has an error, Traefik can remove all your Docker stacks as it is communicating with the Docker swarm manager. We know, that it is only reading information from the Docker swarm service, but it can also send commands there if someone implements it. In our personal opinion, the Docker swarm DNS based service discovery is very useful to retrieve the backend ip address container information. This is, what our binary does.

Another thing is, that Nginx has a lot of configuration possibilities which you might like to have. Therefore this project is based on the Golang text template system. You can write whatever Nginx config you like and replace the backend information with data from the backends.

# Conclusion
This project is far away from being perfect, nor does it reflect perfect written code. It just works and maybe there is someone out there who find it useful. For us it is also nice to have an interface to the [PowerDNS API](https://www.powerdns.com) because an automatic registration of the ip address with the DNS server is extram helpful in a large environment where you have to manage more than two or three services. This project is limited to Docker swarm services!

# Configuration

Here is an example configuration.

## border-controller
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
pdns:
  api_url: https://your.pdns/api/v1/servers/localhost/zones
  api_key: yourapikey
  ip_address: 1.1.1.1
  domain_prefix: funny
  domain_zone: domain.com

```
### debug
If parameter is not provided, default value is set to false.

### check_intervall
If parameter is not provided, default value set to 30 secons. Time duration is in seconds.

### pdns
This configuration section is completely optional.

## nginx.conf
In the ```nginx.conf``` you can do all configuration which is supported by Nginx. There are no restrictions.

```
worker_processes auto;

events {}

http {

    upstream {{.testcontexta.Upstream}} {
       hash $remote_addr;
       {{range $index, $entry := .testcontexta.Servers}} server {{$entry.Server}}:{{$entry.Port}};
       {{end}}
    }

    upstream {{.testcontextb.Upstream}} {
       hash $remote_addr;
       {{range $index, $entry := .testcontexta.Servers}} server {{$entry.Server}}:{{$entry.Port}};
       {{end}}
    }

    server {
        listen 80;
	location / {
        	proxy_pass http://{{.testcontexta.Upstream}};
	}
    }
}
```

# Changelog
You can find the changelog information [here](CHANGELOG.md).

# Version
1.2


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fn0r1sk%2Fborder-controller.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fn0r1sk%2Fborder-controller?ref=badge_large)