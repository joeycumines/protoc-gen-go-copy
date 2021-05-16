# addressbook

Originally based on an example from the official go tutorial.

https://developers.google.com/protocol-buffers/docs/gotutorial

https://developers.google.com/protocol-buffers/docs/reference/go-generated

```bash
# example usage, assumes protoc and other dependencies unrelated to codegen are already installed

cd ~

# if necessary, install the official go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# we need the example proto
git clone https://github.com/joeycumines/protoc-gen-go-copy.git

cd protoc-gen-go-copy

# and we may as well install the go-copy plugin from the same source
go install

# generate the code, into the same directory as the proto file, see also the root readme, and google's reference docs
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-copy_out=. --go-copy_opt=paths=source_relative examples/addressbook/addressbook.proto
```
