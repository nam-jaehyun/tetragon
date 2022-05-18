// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cilium/tetragon/pkg/logger"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TetragonPackageName is the import path for the Tetragon package
var TetragonPackageName = "github.com/cilium/tetragon"

// TetragonApiPackageName is the import path for the code generated package
var TetragonApiPackageName = "api/v1/tetragon"

// TetragonCopyrightHeader is the license header to prepend to all generated files
var TetragonCopyrightHeader = `// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon`

// NewGeneratedFile creates a new codegen pakage and file in the project
func NewGeneratedFile(gen *protogen.Plugin, file *protogen.File, pkg string) *protogen.GeneratedFile {
	importPath := filepath.Join(string(file.GoImportPath), "codegen", pkg)
	pathSuffix := filepath.Base(file.GeneratedFilenamePrefix)
	fileName := filepath.Join(strings.TrimSuffix(file.GeneratedFilenamePrefix, pathSuffix), "codegen", pkg, fmt.Sprintf("%s.pb.go", pkg))
	logger.GetLogger().Infof("%s", fileName)

	g := gen.NewGeneratedFile(fileName, protogen.GoImportPath(importPath))
	g.P(TetragonCopyrightHeader)
	g.P()

	g.P("// Code generated by protoc-gen-go-tetragon. DO NOT EDIT")
	g.P()

	g.P("package ", pkg)
	g.P()

	return g
}

// GoIdent is a convenience helper that returns a qualified go ident as a string for
// a given import package and name
func GoIdent(g *protogen.GeneratedFile, importPath string, name string) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

// TetragonApiIdent is a convenience helper that calls GoIdent with the path to the
// Tetragon API package.
func TetragonApiIdent(g *protogen.GeneratedFile, name string) string {
	return TetragonIdent(g, TetragonApiPackageName, name)
}

// TetragonIdent is a convenience helper that calls GoIdent with the path to the
// Tetragon package.
func TetragonIdent(g *protogen.GeneratedFile, importPath string, name string) string {
	importPath = filepath.Join(TetragonPackageName, importPath)
	return GoIdent(g, importPath, name)
}

// Logger is a convenience helper that generates a call to logger.GetLogger()
func Logger(g *protogen.GeneratedFile) string {
	return fmt.Sprintf("%s()", GoIdent(g, "github.com/cilium/tetragon/pkg/logger", "GetLogger"))
}

// FmtErrorf is a convenience helper that generates a call to fmt.Errorf
func FmtErrorf(g *protogen.GeneratedFile, fmt_ string, args ...string) string {
	args = append([]string{fmt.Sprintf("\"%s\"", fmt_)}, args...)
	return fmt.Sprintf("%s(%s)", GoIdent(g, "fmt", "Errorf"), strings.Join(args, ", "))
}

// GetEvents returns a list of all messages that are events
func GetEvents(file *protogen.File) ([]*protogen.Message, error) {
	var getEventsResponse *protogen.Message
	for _, msg := range file.Messages {
		if msg.GoIdent.GoName == "GetEventsResponse" {
			getEventsResponse = msg
			break
		}
	}
	if getEventsResponse == nil {
		return nil, fmt.Errorf("Unable to find GetEventsResponse message")
	}

	var eventOneof *protogen.Oneof
	for _, oneof := range getEventsResponse.Oneofs {
		if oneof.Desc.Name() == "event" {
			eventOneof = oneof
			break
		}
	}
	if eventOneof == nil {
		return nil, fmt.Errorf("Unable to find GetEventsResponse.event")
	}

	validNames := make(map[string]struct{})
	for _, type_ := range eventOneof.Fields {
		name := strings.TrimPrefix(type_.GoIdent.GoName, "GetEventsResponse_")
		validNames[name] = struct{}{}
	}

	var events []*protogen.Message
	for _, msg := range file.Messages {
		if _, ok := validNames[string(msg.Desc.Name())]; ok {
			events = append(events, msg)
		}
	}

	return events, nil
}

// EventFieldCheck returns true if the event has the field
func EventFieldCheck(msg *protogen.Message, field string) bool {
	if msg.Desc.Fields().ByName(protoreflect.Name(field)) != nil {
		return true
	}

	return false
}

// IsProcessEvent returns true if the message is an Tetragon event that has a process field
func IsProcessEvent(msg *protogen.Message) bool {
	return EventFieldCheck(msg, "process")
}

// IsParentEvent returns true if the message is an Tetragon event that has a parent field
func IsParentEvent(msg *protogen.Message) bool {
	return EventFieldCheck(msg, "parent")
}