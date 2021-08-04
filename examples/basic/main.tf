terraform {
  required_providers {
    mzcloud = {
      source = "MaterializeInc/mzcloud"
    }
  }
}

resource "mzcloud_deployment" "example" {
  size       = "XS"
  mz_version = "v0.8.3"
}
