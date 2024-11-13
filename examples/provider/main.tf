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
    access_token_url = "https://authorization-server.com/token"
    client_id        = "client-id"
    client_secret    = "client-secret"
    scope            = "scope"
    grant_type       = "grant-type"
  }
}
