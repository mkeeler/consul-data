variable "consul_token" {
   type = string
   default = ""
   description = "ACL token to use to access the cluster"
}

variable "consul_address" {
   type = string
   default = ""
   description = "Address of the Consul HTTP API to talk to"
}

variable "data" {
   type = string
   description = "Path to the Consul data generate by consul-data generate all command"
}