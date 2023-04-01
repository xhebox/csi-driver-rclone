job "redis" {
  type = "batch"
  group "redis" {
    volume "redis" {
      type            = "csi"
      read_only       = false
      source          = "pcloud"
      access_mode     = "single-node-writer"
      attachment_mode = "file-system"
    }
    task "redis" {
      driver = "podman"
      volume_mount {
        volume      = "redis"
        destination = "/data"
        read_only   = false
      }
      config {
        image = "docker://redis"
      }
      resources {
        cpu    = 100
        memory = 100
      }
    }
  }
}
