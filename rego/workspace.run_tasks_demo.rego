package workspace.run_tasks_demo

import data.provisioners.local_exec.disallow as local_exec
# import data.foo as foo
# import data.bar as bar

default allow := false

allow := true {
    count(local_exec.deny_provisioners) == 0
    # count(foo.something) == 0
    # count(bar.something) == 0
}

reasons[key] := value {
    count(local_exec.deny_provisioners) > 0
    key := "provisioners.local_exec.disallow"
    value := local_exec.deny_provisioners
}

# reasons[key] := value {
#     count(foo.something) > 0
#     key := "foo"
#     value := foo.something
# }

# reasons[key] := value {
#     count(bar.something) > 0
#     key := "bar"
#     value := bar.something
# }