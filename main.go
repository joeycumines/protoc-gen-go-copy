// Copyright 2021 Joseph Cumines
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command protoc-gen-go-copy is a protoc plugin that generates code to copy messages without reflection.
package main

import (
	"flag"
	"fmt"
	"github.com/jhump/gopoet"
	"github.com/joeycumines/gopoet-protogen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"path"
)

func main() {
	config := DefaultGenerator
	(protogen.Options{ParamFunc: config.NewFlagSet().Set}).Run(func(gen *protogen.Plugin) error {
		config.Plugin = gen
		return config.Generate()
	})
}

var DefaultGenerator = Generator{
	Cache:                   new(gopoet_protogen.Cache),
	GeneratedFilenameSuffix: "_copy.pb.go",
	ShallowCopyMethod:       "Proto_ShallowCopy",
	ShallowCloneMethod:      "Proto_ShallowClone",
}

type (
	Generator struct {
		Plugin                  *protogen.Plugin
		Cache                   Cache
		GeneratedFilenameSuffix string
		ShallowCopyMethod       string
		ShallowCloneMethod      string
	}

	Cache interface {
		AddFile(*protogen.File)
		MessageType(protoreflect.MessageDescriptor) gopoet.TypeName
		MessageFields(*protogen.Message) []gopoet_protogen.Field
	}
)

func (x *Generator) NewFlagSet() *flag.FlagSet {
	var (
		flags   flag.FlagSet
		addFlag = func(v *string, n string) {
			flags.StringVar(v, n, *v, "method name generated for all message types unless set to an empty string")
		}
	)
	addFlag(&x.GeneratedFilenameSuffix, "generated_filename_suffix")
	addFlag(&x.ShallowCopyMethod, "shallow_copy_method")
	addFlag(&x.ShallowCloneMethod, "shallow_clone_method")
	return &flags
}

func (x Generator) Generate() error {
	for _, file := range x.Plugin.Files {
		x.Cache.AddFile(file)

		if !file.Generate {
			continue
		}

		filename := file.GeneratedFilenamePrefix + x.GeneratedFilenameSuffix

		f := gopoet.NewGoFile(path.Base(filename), string(file.GoImportPath), string(file.GoPackageName))
		f.FileComment = fmt.Sprintf("Code generated by protoc-gen-go-copy. DO NOT EDIT.\nsource: %s", file.Desc.Path())

		var genMsg func(msg *protogen.Message)
		genMsg = func(msg *protogen.Message) {
			if msg.Desc.IsMapEntry() {
				return
			}

			var (
				t      = x.Cache.MessageType(msg.Desc)
				fields = x.Cache.MessageFields(msg)
			)

			if x.ShallowCopyMethod != `` {
				elem := gopoet.NewMethod(gopoet.NewPointerReceiverForType(`x`, t), x.ShallowCopyMethod)
				elem.SetComment(fmt.Sprintf(`%s copies fields, from v to the receiver, using field getters.
Note that v is of an arbitrary type, which may implement any number of the
field getters, which are defined as any methods of the same signature as those
generated for the receiver type, with a name starting with Get.`, x.ShallowCopyMethod))
				elem.AddArg(`v`, gopoet.InterfaceType(nil))
				if len(fields) != 0 {
					elem.Printlnf(`switch v := v.(type) {`)
					elem.Printlnf(`case %s:`, gopoet.PointerType(t))
					for _, field := range fields {
						elem.Printlnf(`x.%s = v.%s()`, field.Name(), field.Getter().Name)
					}
					elem.Println(`default:`)
					for _, field := range fields {
						elem.Printlnf(`if v, ok := v.(%s); ok {`, gopoet.InterfaceType(nil, field.Getter()))
						elem.Printlnf(`x.%s = v.%s()`, field.Name(), field.Getter().Name)
						if field.OneOf() != nil {
							elem.Println(`} else {`)
							elem.Println(`func() {`)
							for _, oneOfField := range field.OneOfFields() {
								elem.Printlnf(`if v, ok := v.(%s); ok {`, gopoet.InterfaceType(nil, oneOfField.Getter))
								if fieldType := oneOfField.Getter.Signature.Results[0].Type; fieldType.Kind() == gopoet.KindSlice {
									// special case to handle field of types `bytes` - can't do slice == slice
									elem.Printlnf(`if v := v.%s(); v != nil {`, oneOfField.Getter.Name)
								} else {
									elem.Printlnf(`var defaultValue %s`, fieldType)
									elem.Printlnf(`if v := v.%s(); v != defaultValue {`, oneOfField.Getter.Name)
								}
								elem.Printlnf(`x.%s = &%s{%s: v}`, field.Name(), oneOfField.Type, oneOfField.Field.GoName)
								elem.Println(`return`)
								elem.Println(`}`)
								elem.Println(`}`)
							}
							elem.Println(`}()`)
						}
						elem.Println(`}`)
					}
					elem.Println(`}`)
				}
				f.AddElement(elem)
			}

			if x.ShallowCloneMethod != `` {
				elem := gopoet.NewMethod(gopoet.NewPointerReceiverForType(`x`, t), x.ShallowCloneMethod)
				elem.SetComment(fmt.Sprintf(`%s returns a shallow copy of the receiver or nil if it's nil.`, x.ShallowCloneMethod))
				elem.AddResult(`c`, gopoet.PointerType(t))
				elem.Println(`if x != nil {`)
				elem.Printlnf(`c = new(%s)`, t)
				for _, field := range fields {
					elem.Printlnf(`c.%s = x.%s`, field.Name(), field.Name())
				}
				elem.Println(`}`)
				elem.Println(`return`)
				f.AddElement(elem)
			}

			for _, msg := range msg.Messages {
				genMsg(msg)
			}
		}
		for _, msg := range file.Messages {
			genMsg(msg)
		}

		if err := gopoet.WriteGoFile(x.Plugin.NewGeneratedFile(filename, file.GoImportPath), f); err != nil {
			switch e := err.(type) {
			case *gopoet.FormatError:
				return fmt.Errorf("%s: error in generated Go code: %w:\n%s", file.Desc.Path(), e, e.Unformatted)
			default:
				return fmt.Errorf("%s: %w", file.Desc.Path(), e)
			}
		}
	}

	return nil
}
