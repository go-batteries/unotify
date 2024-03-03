provisioner "soc" {
  aliasmapper "soc" {
    aliases = {
      to_do: "To Do",
      in_progress: "In Progress",
      done: "Done"
    }
  }

  statemachine "soc" {
    initial = "to_do"

    state "to_do" "next" {
      transition = "in_progress"
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
}

provisioner "devhop" {
  aliasmapper "devhop" {
    aliases = {
      to_do: "To Do",
      in_progress: "In Progress",
      done: "Done"
    }
  }

  statemachine "devhop" {
    initial = "to_do"

    state "to_do" "next" {
      transition = "in_progress"
    }

    state "in_progress" "next" {
      transition = "done"
    }

    state "done" "next" {
      transition = "<end>"
    }
  }
}
