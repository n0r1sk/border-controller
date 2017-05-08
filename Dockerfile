FROM ubuntu:16.04

MAINTAINER Mario Kleinsasser "mario.kleinsasser@gmail.com"
MAINTAINER Bernhard Rausch "rausch.bernhard@gmail.com"

RUN apt-get update && apt-get -y install wget

RUN mkdir /data
RUN wget -O /data/traefik https://github.com/containous/traefik/releases/download/v1.2.3/traefik_linux-amd64 
RUN chmod 755 /data/traefik

ADD ingress-controller /data/ingress-controller
RUN chmod 755 /data/ingress-controller

ADD ingress-controller-config.tpl /config/ingress-controller-config.tpl

CMD ["/data/ingress-controller"]
