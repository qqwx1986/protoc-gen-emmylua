// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package genlua

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
	"strings"
	"time"
)

// SupportedFeatures reports the set of supported protobuf language features.
var SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

// GenerateVersionMarkers specifies whether to generate version markers.
var GenerateVersionMarkers = true

// GenerateFile generates the contents of a .pb.go file.
func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	path := file.Desc.Path()
	var filename = ""
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			filename = path[:i] + ".lua"
			break
		}
	}
	if filename == "" {
		filename = path + ".lua"
	}
	g := gen.NewGeneratedFile(filename, "")
	f := newFileInfo(file)

	genGeneratedHeader(gen, g, f)

	for _, enum := range f.allEnums {
		genEnum(g, f, enum)
	}
	for _, message := range f.allMessages {
		genMessage(g, f, message)
	}
	g.P("return proto")
	return g
}

func genGeneratedHeader(gen *protogen.Plugin, g *protogen.GeneratedFile, f *fileInfo) {
	g.P("-- Code generated by protoc-gen-emmylua. DO NOT EDIT.")

	if GenerateVersionMarkers {
		g.P("-- versions:")
		protocGenGoVersion := "v1.0"
		protocVersion := "(unknown)"
		if v := gen.Request.GetCompilerVersion(); v != nil {
			protocVersion = fmt.Sprintf("v%v.%v.%v", v.GetMajor(), v.GetMinor(), v.GetPatch())
		}
		g.P("-- \tprotoc-gen-emmylua ", protocGenGoVersion)
		g.P("-- \tprotoc        ", protocVersion)
		g.P("-- \ttime          ", time.Now().Format(time.DateTime))
	}

	if f.Proto.GetOptions().GetDeprecated() {
		g.P("-- ", f.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("-- source: ", f.Desc.Path())
	}
	g.P()
	g.P("local proto = {}")
	g.P()
}

func genEnum(g *protogen.GeneratedFile, f *fileInfo, e *enumInfo) {
	enumName := e.GoIdent.GoName
	comments := trailingComment(e.Comments.Leading)
	g.P("---@class ", enumName, " @", comments)
	for _, value := range e.Values {
		g.P("---@field public ", value.Desc.Name(), " integer|string @", value.Desc.Number(), " ", trailingComment(value.Comments.Trailing))
	}
	g.P()
	g.P("---@type ", enumName)
	g.P("proto.", enumName, " = {")
	for _, value := range e.Values {
		g.P("\t", value.Desc.Name(), " = ", value.Desc.Number(), ",")
	}
	g.P("}")
	g.P()

	g.P("---@type ", enumName)
	g.P("proto.", enumName, "Str = {")
	for _, value := range e.Values {
		g.P("\t", value.Desc.Name(), " = \"", value.Desc.Name(), "\",")
	}
	g.P("}")
	g.P()
}

func genMessage(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	if m.Desc.IsMapEntry() {
		return
	}
	comments := trailingComment(m.Comments.Leading)
	g.P("---@class ", m.GoIdent.GoName, " @", comments)
	genMessageFields(g, f, m)
	g.P()
	g.P("proto.", m.GoIdent.GoName, " = ", "\"", m.GoIdent.GoName, "\"")
}

func genMessageFields(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	sf := f.allMessageFieldsByPtr[m]
	for _, field := range m.Fields {
		genMessageField(g, f, m, field, sf)
	}
}

func genMessageField(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, field *protogen.Field, sf *structFields) {
	goType := fieldGoType(g, f, field)
	name := field.GoName
	comments := trailingComment(field.Comments.Trailing)
	if comments != "" {
		comments = "@" + comments
	}
	g.P("---@field ", name, " ", goType, " ", comments)
	sf.append(field.GoName)
}

// fieldGoType returns the lua type used for a field.
func fieldGoType(g *protogen.GeneratedFile, f *fileInfo, field *protogen.Field) (goType string) {
	if field.Desc.IsWeak() {
		return ""
	}

	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "boolean"
	case protoreflect.EnumKind:
		goType = field.Enum.GoIdent.GoName
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "integer"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "integer"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "integer"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "integer"
	case protoreflect.FloatKind:
		goType = "float"
	case protoreflect.DoubleKind:
		goType = "number"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "any"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		goType = field.Message.GoIdent.GoName
	}
	switch {
	case field.Desc.IsList():
		return goType + "[]"
	case field.Desc.IsMap():
		keyType := fieldGoType(g, f, field.Message.Fields[0])
		valType := fieldGoType(g, f, field.Message.Fields[1])
		return fmt.Sprintf("table<%v,%v>", keyType, valType)
	}
	return goType
}

// trailingComment is like protogen.Comments, but lacks a trailing newline.
type trailingComment protogen.Comments

func (c trailingComment) String() string {
	s := strings.TrimSuffix(protogen.Comments(c).String(), "\n")
	if strings.Contains(s, "\n") {
		// We don't support multi-lined trailing comments as it is unclear
		// how to best render them in the generated code.
		return ""
	}
	if len(s) > 2 {
		return s[2:]
	}
	return s
}
