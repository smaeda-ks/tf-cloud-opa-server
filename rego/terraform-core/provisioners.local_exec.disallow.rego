package provisioners.local_exec.disallow

# ref: https://engineering.mercari.com/en/blog/entry/20220519-terraform-ci-code-execution-restrictions/

values_with_path(value, path) = r {
    # skip 'root_module.module_calls' and 'module.module_calls'
    module_path = [m | m = path[i]; (i+1) % 3 == 0; (i+1) <= count(path) - 2; i > 0]
    r = [
    {
        "path": concat(".", full_path), 
        "value":val
    } | val := value[i]; address := value[i].address; full_path := array.concat(module_path,[address])]
}

resources[r] {
    some path, value
    # Walk over the JSON tree and check root and child modules
    walk(input.configuration, [path, value])
    # Look for resources in the current value based on path
    rs := module_resources(path, value)
    # Aggregate them into "resources"
    r := rs[_]
}

# Variant to match root_module resources
module_resources(path, value) = rs {
    # Where the path is [..., "root_module", "resources"]
    reverse_index(path, 1) == "resources"
    reverse_index(path, 2) == "root_module"
    rs := values_with_path(value, path)
}

# Variant to match child_modules resources
module_resources(path, value) = rs {
    # match [..., "module_calls", i, "module", "resources"]
    reverse_index(path, 1) == "resources"
    reverse_index(path, 2) == "module"
    reverse_index(path, 4) == "module_calls"
    rs := values_with_path(value, path)
}

reverse_index(path, idx) = value {
    value := path[count(path) - idx]
}

deny_provisioners[msg] {
    count(resources[i].value.provisioners) > 0
    msg = sprintf("Provisioner found at path: '%s'!", [resources[i].path])
}
