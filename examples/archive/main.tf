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
  type = "xxx"
  out_mode = null

  exclude_list = ["/path/example/dir", "/path/example.txt"]

  resolve_symlink = null

  file {
    path = "../path/file.txt"
  }

  file {
    path = "../path/file.txt"
  }

  dir {
    path = "../path/dir"
  }

  content {
    src = base64encode("content")
    file_path = "/path/store/src/content.txt"
  }

}
