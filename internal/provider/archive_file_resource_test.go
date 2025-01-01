package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Wa4h1h/terraform-provider-archiver/internal/archive"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestArchiveFileResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "archiver_file" "test" {
  name = "test.zip"
  type = "zip"

  out_mode="777"

  exclude_list=["../../internal/archive/archiver.go"]

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
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("archiver_file.test", "name", "test.zip"),
					resource.TestCheckResourceAttr("archiver_file.test", "type", "zip"),
					resource.TestCheckResourceAttr("archiver_file.test", "out_mode", "777"),
					resource.TestCheckResourceAttr("archiver_file.test", "exclude_list.0",
						"../../internal/archive/archiver.go"),
					resource.TestCheckResourceAttr("archiver_file.test", "file.0.path",
						"../../internal/archive/archiver.go"),
					resource.TestCheckResourceAttr("archiver_file.test", "file.1.path",
						"../../internal/provider/provider.go"),
					resource.TestCheckResourceAttr("archiver_file.test", "dir.0.path",
						"../../.github"),
					resource.TestCheckResourceAttr("archiver_file.test", "content.0.src", "Y29udGVudA=="),
					resource.TestCheckResourceAttr("archiver_file.test", "content.0.file_path", "content.txt"),
					func(s *terraform.State) error {
						for _, r := range s.Modules[0].Resources {
							absPath := r.Primary.Attributes["abs_path"]
							a, err := filepath.Abs("test.zip")
							if err != nil {
								return err
							}

							if a != absPath {
								return fmt.Errorf("expected output abs path %s but got %s", a, absPath)
							}

							md5, sha256, err := archive.Checksums(a)
							if err != nil {
								return err
							}

							md5Out := r.Primary.Attributes["md5"]
							if md5 != md5Out {
								return fmt.Errorf("expected output md5 %s but got %s", md5Out, md5)
							}

							sha256Out := r.Primary.Attributes["sha256"]
							if sha256 != sha256Out {
								return fmt.Errorf("expected output sha256 %s but got %s", sha256Out, sha256)
							}

							b, err := os.ReadFile(absPath)
							if err != nil {
								return err
							}

							sizeOut := r.Primary.Attributes["size"]
							size := fmt.Sprintf("%d", len(b))
							if size != sizeOut {
								return fmt.Errorf("expected output size %s but got %s", sizeOut, size)
							}

						}

						return nil
					},
				),
			},
			{
				Config: providerConfig + `
resource "archiver_file" "test" {
  name = "example.zip"
  type = "tar.gz"

  out_mode="666"

  exclude_list=["../../internal/archive/archiver.go"]

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
`, Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("archiver_file.test", "name", "example.zip"),
					resource.TestCheckResourceAttr("archiver_file.test", "type", "tar.gz"),
					resource.TestCheckResourceAttr("archiver_file.test", "out_mode", "666")),
			},
		},
	})
}
