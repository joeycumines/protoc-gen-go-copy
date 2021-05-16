# protoc-gen-go-copy

Command protoc-gen-go-copy is a protoc plugin that generates code to copy messages without reflection.

## Features

See also [examples/addressbook](examples/addressbook).

```go
// ShallowCopy copies fields, from v to the receiver, using field getters.
// Note that v is of an arbitrary type, which may implement any number of the
// field getters, which are defined as any methods of the same signature as those
// generated for the receiver type, with a name starting with Get.
func (x *Policy) ShallowCopy(v interface{}) {
	switch v := v.(type) {
	case *Policy:
		x.Name = v.GetName()
		x.Id = v.GetId()
		x.Spec = v.GetSpec()
	default:
		if v, ok := v.(interface{ GetName() string }); ok {
			x.Name = v.GetName()
		}
		if v, ok := v.(interface{ GetId() rune }); ok {
			x.Id = v.GetId()
		}
		if v, ok := v.(interface{ GetSpec() isPolicy_Spec }); ok {
			x.Spec = v.GetSpec()
		} else {
			func() {
				if v, ok := v.(interface{ GetDeletePerson() *Policy_DeletePerson }); ok {
					if v := v.GetDeletePerson(); v != nil {
						x.Spec = &Policy_DeletePerson_{DeletePerson: v}
						return
					}
				}
				if v, ok := v.(interface {
					GetDefaultPolicyRejectBlocked() *DefaultPolicy_RejectBlocked
				}); ok {
					if v := v.GetDefaultPolicyRejectBlocked(); v != nil {
						x.Spec = &Policy_DefaultPolicyRejectBlocked{DefaultPolicyRejectBlocked: v}
						return
					}
				}
				if v, ok := v.(interface{ GetAny() *anypb.Any }); ok {
					if v := v.GetAny(); v != nil {
						x.Spec = &Policy_Any{Any: v}
						return
					}
				}
			}()
		}
	}
}

// ShallowClone returns a shallow copy of the receiver or nil if it's nil.
func (x *Policy) ShallowClone() (c *Policy) {
	if x != nil {
		c = new(Policy)
		c.Name = x.Name
		c.Id = x.Id
		c.Spec = x.Spec
	}
	return
}
```

## Usage

```bash
# install the plugin
go install github.com/joeycumines/protoc-gen-go-copy@latest

# example protoc usage based on https://developers.google.com/protocol-buffers/docs/reference/go-generated
protoc --proto_path=src --go_out=out --go_opt=paths=source_relative --go-copy_out=out --go-copy_opt=paths=source_relative foo.proto bar/baz.proto
```

This plugin generates methods for message types, written to `*_copy.pb.go` file(s). This suffix may be configured using
the `generated_filename_suffix` option. These files are intended to sit alongside the usual `*.pb.go` files (and
`*_grpc.pb.go` files, if you're using gRPC).

All standard plugin options are supported. This includes the `paths` option (see above example), and others, including
[go package overrides](https://developers.google.com/protocol-buffers/docs/reference/go-generated#package).

Additional plugin options may be provided in the
[usual](https://github.com/protocolbuffers/protobuf-go/blob/0e358a402f994eaf96a258b7c5c5b3317d4575aa/compiler/protogen/protogen.go#L129)
way, using the format `--<plugin>_out=<param1>=<value1>,<param2>=<value2>:<output_directory>`. For example,
`--go-copy_out=shallow_copy_method=,shallow_clone_method=:out` would have the effect of deleting any existing source code, that was generated
using this plugin.

| Option | Description |
| --- | --- |
| generated_filename_suffix | filename suffix for all generated files, defaults to `_copy.pb.go` |
| shallow_copy_method | method name generated for all message types unless set to an empty string, defaults to `ShallowCopy` |
| shallow_clone_method | method name generated for all message types unless set to an empty string, defaults to `ShallowClone` |