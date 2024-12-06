terraform {
  required_providers {
    tools = {
      source = "registry.terraform.io/Wa4h1h/tools"
    }
  }
}

// http object is optional
provider "tools" {
  http = {
    hostname="hostname"
    headers={
      "Content-Type"="application/json"
    }
    // in seconds
    timeout=1
  }
}
