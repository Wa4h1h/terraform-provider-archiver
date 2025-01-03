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

  out_mode = "777"

  exclude_list = ["../../example/example.txt", ".../../../dir"]

  file {
    path = "../../xx/yy.txt"
  }

  dir {
    path = "../../dir"
  }

  content {
    src       = base64encode("content")
    file_path = "content.txt"
  }
}

output "md5" {
  value = archiver_file.archive.md5
}

output "sha256" {
  value = archiver_file.archive.sha256
}

output "size" {
  value = archiver_file.archive.size
}
