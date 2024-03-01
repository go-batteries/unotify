Handling jira states.

Conventions:

`ISSUE_ID` is made of `PROJECT + "-" + ID` .

Store states per PROJECT basis. We will just start with an YAML file,
but configuration can be stored in cache as well. Preloaded.

[SOC] = {
  ["To Do"] = {
    ["ToStates"]  = { "In Progress", "Done" }
  },

  ["In Progress"] = {
    ["ToStates"] = { "Done" }
  }

  ["Done"] = {
    ["ToStates"] = {}
  }
}

Could we turn this into a DSL


