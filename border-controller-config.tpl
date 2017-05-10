worker_processes auto;

events {
  worker_connections  4096;
}

http {
    ssl_session_cache	shared:SSL:10m;
    ssl_session_timeout	10m;
    ssl_prefer_server_ciphers on;

    upstream upstreams {
        hash $remote_addr;
	{{range $index, $entry := .}} server {{$entry.Node}}:{{$entry.Port}};{{end}}
    }

    server {
	listen 443 ssl;
	
	ssl_certificate /etc/nginx/ssl/certificate.pem;
	ssl_certificate_key /etc/nginx/ssl/private.key;
	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA:ECDHE-RSA-AES128-SHA:AES128-SHA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!MD5:!PSK:!RC4;

	location / {
        	proxy_pass http://upstreams;
	}
    }

    server {
	listen 80;
	return 302 https://$server_name$request_uri;
    }
}
