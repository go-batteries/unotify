# Readme

Describes the DSL for defining state machines.

Requirements:

- hcl

The example is going to use Jira, but probably can be extended
for other use cases.

- Each Jira Project is a `provisioner`
  - Small case project name is used instead of original project name
- Each Column in the Board a `state`
  - The slug of these column names are used as aliases for `state`
- Issue ID should be of the form `{PROJECT_NAME}-{IssueID}`
- Statemachine should have an ending state, characterized
  with a special token `<end>`

Example Scenario:

A jira project named `DEVOPS` has 3 columns in the board:

- To do
- In progress
- In QA
- Done

Provisioner definition would be:

```hcl
provisioner "devops" {

}
```

Adding an alias for each of those columns:

```hcl
provisioner "devops" {
    aliasmapper "devops" {
        aliases = {
            to_do: "To do",
            in_progress: "In Progress",
            in_qa: "In QA",
            done: "Done"
        }
    }
}
```

Assuming, the following state transitions:

- To do -> In progress
- In progress -> In QA
- In QA -> In progress
- In QA -> Done

```hcl
provisioner "devops" {
    aliasmapper "devops" {
        aliases = {
            to_do: "To do",
            in_progress: "In Progress",
            in_qa: "In QA",
            done: "Done"
        }
    }
}

statemachine "devops" {
    initial = "to_do"

    state "to_do" "next" {
      transition = "in_progress"
    }

    state "in_progress" "next" {
      transition = "in_qa"
    } 

    state "in_qa" "prev" {
      transition = "in_progress"
    }

    state "in_qa" "next" {
      transition = "done"
    }

    state "done" "next" {
      transition = "<end>"
    }
}
```

You can have one statemachine per provisioner.

**N.B.:** Need to set the active provisoner names in `config/app.env`
