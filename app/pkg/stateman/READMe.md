## StateMachine

**Deterministic FiniteStateMachine**, is the word they say to describe:
`Only a specific event action on a jira state can lead to one outcome`

This is used in Jira Workers to transition ticket states. Each column in a jira
board, represents a `state`. The `state` has an `id` and `name` associated with it.

**Sidenote**:

> The names of the states by jira is a bit weird, the api returns names in
`British Case`. So `To Do` instead of `TODO` as we see in `UI`. I am gonna just
keep the `state names` as is, coz:

_Relying on 3rd party APIs to not change is a fucking joke, So any transformation is prone to break in future_.



Instead there will be an internal  `API` which lists the available `states` for
a given _Project_ or _Issue-ID_.(_`Issue-ID` is basically `PROJECT_NAME-<int>`_)



The frozen state machine config is provided as an `HCL` file. The structure
looks as such:

```hcl
statemachine "soc" {
  states = ["to_do", "in_progress", "done"]
  entrypoint = "to_do" // optional

  state "to_do" "prev" {
    transition = "in_progress"
  }

  state "in_progress" "next" {
    transition = "done"
  } 

  state "in_progress" "prev" {
    transition = "to_do"
  }

  state "done" "next"{
    transition = "<end>"
  }
}
```

You will also need to create an `aliasmapper` to map a better slug to the
provided `state name` by jira. Because, you are a bad conductor of
stupidity.


```hcl
aliasmapper "soc" {
  aliases = {
    to_do: "To Do",
    in_progress: "In Progress",
    done: "Done"
  }
}
```


**Why HCL?**

It looks good. Its descriptive. Also, different jira boards, can have different
`states`, most may be same, But its better to be specific, than sprinkling magic.


```
AliasInverseMap {
  "To Do": "to_do",
}

State {
  Alias string
  Name string
  Event string
  Transition string
}

NextState(currState State, event string, inverted bool) {
  if inverted {
    currState = AliasInverseMap[currState]
  }

  transition, ok = stateMap[currState][event]
  transition == "<end>"
  
}
```

```


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


