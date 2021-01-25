locals {
   data = jsondecode(file(var.data))
   
   kv = lookup(local.data, "KV", {})
   
   nodes = lookup(local.data, "Catalog", [])
   
   services = flatten([
      for node in local.nodes:
         [for service in lookup(node, "Services", []):
            {"service": service, "node": node}]
   ])
   
}

resource "consul_keys" "kv_entries" {
   for_each = local.kv
   
   datacenter = lookup(each.value, "Datacenter", "")
   token = lookup(each.value, "Token", "")
   namespace = lookup(each.value, "Namespace", "")
   
   key {
      path = each.key
      value = each.value.Value
      flags = lookup(each.value, "Flags", 0)      
   }   
}

resource "consul_node" "nodes" {
   for_each = local.nodes
   
   datacenter = lookup(each.value, "Datacenter", "")
   address = lookup(each.value, "Address", "")
   name = each.value.Name
   meta = lookup(each.value.Meta, "Meta", {})
}