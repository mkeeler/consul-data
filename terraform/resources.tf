resource "consul_keys" "kv_entries" {
   for_each = jsondecode(file(var.kv_data))
   
   datacenter = lookup(each.value, "Datacenter", "")
   token = lookup(each.value, "Token", "")
   namespace = lookup(each.value, "Namespace", "")
   
   key {
      path = each.key
      value = each.value.Value
      flags = lookup(each.value, "Flags", 0)      
   }   
}