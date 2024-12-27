terraform {
  required_providers {
    archiver = {
      source = "registry.terraform.io/Wa4h1h/archiver"
    }
  }
}

provider "archiver" {}

resource "archiver_file" "archive" {
  name = "example.zip"
  type = "zip"

  file {
    path = "../../internal/archive/archiver.go"
  }

  file {
    path = "../../internal/provider/provider.go"
  }

  dir {
    path = "../../.github"
  }

  content {
    src = base64encode("content")
    file_path = "content.txt"
  }

}
