variable "proxmox_api_url" {
  type        = string
  description = "https://<proxmox-ip>:8006/api2/json"
}

variable "proxmox_api_token_id" {
  type        = string
  sensitive   = true
}

variable "proxmox_api_token_secret" {
  type        = string
  sensitive   = true
}

variable "target_node" {
  type        = string
  default     = "pve"
  description = "The Proxmox host node (host1 or host2)"
}

variable "vm_name" {
  type        = string
  description = "Unique name for the environment VM"
}

variable "vm_id" {
  type        = number
  description = "Proxmox VM ID"
}

variable "clone_template" {
  type        = string
  default     = "ubuntu-2204-cloudinit-template"
  description = "The base template to clone from"
}

variable "cores" {
  type    = number
  default = 2
}

variable "memory" {
  type    = number
  default = 2048
}

variable "disk_size" {
  type    = string
  default = "20G"
}
