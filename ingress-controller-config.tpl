worker_processes auto;

http {

    upstream upstreams {
	{{range .}}
	server {{.Hostname}}:{{.Port}}; {{end}}

	sticky learn
          create=$upstream_cookie_examplecookie
          lookup=$cookie_examplecookie
          zone=client_sessions:1m;
    }

    server {
        listen 80;
        location / {
                proxy_pass http://upstreams;
        }
    }
}

