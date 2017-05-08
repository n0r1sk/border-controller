FROM nginx:1.13.0

MAINTAINER Mario Kleinsasser "mario.kleinsasser@gmail.com"
MAINTAINER Bernhard Rausch "rausch.bernhard@gmail.com"

ADD ingress-controller /data/ingress-controller
RUN chmod 755 /data/ingress-controller

ADD ingress-controller-config.tpl /config/ingress-controller-config.tpl

CMD ["/data/ingress-controller"]
