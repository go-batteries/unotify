## Finite StateMachine
 








## HCL for state machine config

HCL to golang mapping. [Example](https://pkg.go.dev/github.com/hashicorp/hcl2/gohcl#example-EncodeIntoBody) here should be explanatory.


```hcl
io_mode = "async"

service "http" "web_proxy" {
  listen_addr = "127.0.0.1:8080"
  
  process "main" {
    command = ["/usr/local/bin/awesome-app", "server"]
  }

  process "mgmt" {
    command = ["/usr/local/bin/awesome-app", "mgmt"]
  }
}
```


Syntax: `ThingType string `hcl:"thing_type,attr"``

```go
type ServiceConfig struct {
  Type       string `hcl:"type,label"`
  Name       string `hcl:"name,label"`
  ListenAddr string `hcl:"listen_addr"`
}
type Config struct {
  IOMode   string          `hcl:"io_mode"`
  Services []ServiceConfig `hcl:"service,block"`
}
```

In this:

- `service` is `,block` in `HCL` tag, attr=block
- `"http" and "web_proxy"`are attr `,label`
- rest of it is same as json

