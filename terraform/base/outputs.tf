output "vm_ip" {
  value       = proxmox_vm_qemu.dynamic_vm.default_ipv4_address
  description = "The assigned IP address of the dynamically provisioned VM"
}

output "vm_status" {
  value = proxmox_vm_qemu.dynamic_vm.vm_state
}
