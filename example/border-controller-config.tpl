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
    }
}
