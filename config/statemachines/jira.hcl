aliasmapper "soc" {
  aliases = {
    to_do: "To Do",
    in_progress: "In Progress",
    done: "Done"
  }
}

statemachine "soc" {
  states = ["to_do", "in_progress", "done"]
  initial = "to_do"

  state "to_do" "next" {
    transition = "in_progress"
  }

  state "to_do" "up" {
    transition = "done"
  }

  state "in_progress" "next" {
    transition = "done"
  } 

  state "in_progress" "prev" {
    transition = "to_do"
  }

  state "done" "next" {
    transition = "<end>"
  }
}
