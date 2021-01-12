terraform {
   required_providers {
      consul = {
         source = "hashicorp/consul"
      }
   }
}

provider "consul" {
  address = var.consul_address
  token   = var.consul_token
}
