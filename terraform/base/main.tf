resource "proxmox_vm_qemu" "dynamic_vm" {
  name        = var.vm_name
  target_node = var.target_node
  vmid        = var.vm_id
  
  clone       = var.clone_template
  full_clone  = true
  
  cores       = var.cores
  sockets     = 1
  memory      = var.memory
  
  network {
    model  = "virtio"
    bridge = "vmbr0"
  }

  disk {
    type    = "scsi"
    storage = "local-lvm"
    size    = var.disk_size
  }

  # Cloud-Init configuration
  os_type   = "cloud-init"
  ipconfig0 = "ip=dhcp" # Or static if managed by OpsPilot
  
  sshkeys = <<EOF
  ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... (OpsPilot Public Key)
  EOF

  lifecycle {
    ignore_changes = [
      network,
    ]
  }
}
