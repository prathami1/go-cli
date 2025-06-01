# Variables for deployment configuration

variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "region" {
  description = "Deployment region"
  type        = string
}

variable "app_type" {
  description = "Type of application"
  type        = string
}

variable "enable_database" {
  description = "Whether to enable database"
  type        = bool
  default     = false
}

variable "enable_storage" {
  description = "Whether to enable storage"
  type        = bool
  default     = false
}

variable "enable_load_balancer" {
  description = "Whether to enable load balancer"
  type        = bool
  default     = false
}
