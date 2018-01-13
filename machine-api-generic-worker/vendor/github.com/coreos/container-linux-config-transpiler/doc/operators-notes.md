# Operator Notes

## Type GUID aliases

The Config Transpiler supports several aliases for GPT partition type GUIDs. They are as follows:

| Alias Name | Resolved Type GUID |
| -- | -- |
| `raid_containing_root` | `be9067b9-ea49-4f15-b4f6-f36f8c9e1818` | 
| `linux_filesystem_data` | `0fc63daf-8483-4772-8e79-3d69d8477de4` |
| `swap_partition` | `0657fd6d-a4ab-43c4-84e5-0933c84b4f4f` |
| `raid_partition` | `a19d880f-05fc-4d3b-a006-743f0f84911e` |

See the [Root Filesystem Placement](https://coreos.com/os/docs/latest/root-filesystem-placement.html) documentation for when to use `raid_containing_root`.
