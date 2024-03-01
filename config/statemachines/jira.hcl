statemachine "soc" {
  states = [ "to_do", "in_progress", "done" ]

  state "to_do" {
    alias = "To Do"
    event = "markup"
    transitions = ["in_progress"]
  }

  state "in_progress" {
    alias = "In Progress"
    event = "markdown"
    transitions = ["to_do"]
  } 

  state "in_progress" {
    alias = "In Progress"
    event = "markup"
    transitions = ["done"]
  }

  state "done" {
    alias = "Done"
    event = "markup"
    transitions = ["-"]
  }
}

