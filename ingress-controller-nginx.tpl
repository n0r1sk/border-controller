worker_processes auto;

events {
  worker_connections  4096;  ## Default: 1024
}

http {

    upstream upstreams {
        {{range .}}
	server {{.Hostname}}:{{.Port}}; {{end}}
    }

    server {
        listen 80;
    }
}
