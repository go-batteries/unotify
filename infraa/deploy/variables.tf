variable AWS_REGION {
  type = string
  default = "ap-south-1"
}

variable AWS_ACCOUNT {
  type = string
}

variable AWS_PROFILE {
  type = string
  default = "default"
}

variable ENVIRONMENT {
  type = string
  default = "prod"
}

variable APP_PORT {
  type = number
  default = 9091
}

variable WORKER_PORT {
  type = number
  default = 9093
}

variable ECS_KEY_NAME {
  type = string
  default="dash-tf-key-pair"
}

variable VPC_NAME {
  type = string
  default = "DashDotDashVPC"
}


variable ECS_CLUSTER_NAME {
  type= string
  default = "DashDotDashCluster"
}

variable APP_NAME {
  type = string
  default = "dashdotdash"
}

variable APP_VERSION {
  type = string
  default = "38a51e4c"
}


variable ATLASSIAN_API_KEY {
  type = string
}

variable ATLASSIAN_EMAIL {
  type = string
}
