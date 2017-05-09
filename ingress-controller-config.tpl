[entryPoints]
  [entryPoints.http]
  address = ":80"

[file]
watch = true

[web]
address = ":9191"
[backends]
  [backends.backend1]
    [backends.backend1.loadbalancer]
      sticky = true
      method = "drr"
    {{range $index, $entry := .}} 
    [backends.backend1.servers.server{{$index}}]
      url = "http://{{$entry.Hostname}}:{{$entry.Port}}" 
      weight = 1{{end}}

[frontends]
  [frontends.frontend1]
  entrypoints = ["http"]
  backend = "backend1"
