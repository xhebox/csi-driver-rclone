type = "csi"
plugin_id = "csi-rclone"
name = "pcloud"
id = "pcloud"
external_id = "pcloud"

capability {
        access_mode = "single-node-writer"
        attachment_mode = "file-system"
}

context {
        type = "pcloud"
        path = "/"
        pcloud-token = <<EOF
xxxxxx
        EOF
        vfs-cache-mode = "full"
}
