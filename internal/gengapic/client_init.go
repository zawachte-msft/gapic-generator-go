// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gengapic

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/googleapis/gapic-generator-go/internal/pbinfo"
	golangGen "github.com/golang/protobuf/protoc-gen-go/generator"
)


func durationToMillis(d *duration.Duration) int64 {
	return d.GetSeconds()*1000 + int64(d.GetNanos()/1000000)
}

func (g *generator) clientInit(serv *descriptor.ServiceDescriptorProto, servName string) error {
	p := g.printf

	imp, err := g.descInfo.ImportSpec(serv)
	if err != nil {
		return err
	}


	// client struct
	{
		p("// %sClient is a client for interacting with %s.", servName, g.apiName)
		p("//")
		p("// Methods, except Close, may be called concurrently. However, fields must not be modified concurrently with method calls.")
		p("type %sClient struct {", servName)

		p("// The connection to the service.")
		p("conn *grpc.ClientConn")
		p("")

		p("// The gRPC API client.")
		p("%s %s.%sClient", grpcClientField(servName), imp.Name, serv.GetName())
		p("")

		p("}")
		p("")

		g.imports[pbinfo.ImportSpec{Path: "google.golang.org/grpc"}] = true

	}

	// Client constructor
	{
		clientName := camelToSnake(serv.GetName())
		clientName = strings.Replace(clientName, "_", " ", -1)

		p("// New%sClient creates a new %s client.", servName, clientName)
		p("//")
		g.comment(g.comments[serv])
		p("func New%[1]sClient(serverAddress *string, authorizer auth.Authorizer) (*%[1]sClient, error) {", servName)
		p("  opts := helper.GetDefaultDialOption(authorizer)")
		p("  conn, err := grpc.Dial(helper.GetServerEndpoint(serverAddress), opts...)")
		p("  if err != nil {")
		p("    klog.Fatalf(\"Unable to get %[1]sClient. Failed to dial the server\")", servName)
		p("  }")
		p("  c := &%sClient{", servName)
		p("    conn:        conn,")
		p("    %s: %s.New%sClient(conn),", grpcClientField(servName), imp.Name, serv.GetName())
		p("  }")
		p("")

		p("  return c, nil")
		p("}")
		p("")

		g.imports[pbinfo.ImportSpec{Path: "github.com/microsoft/wssd-sdk-for-go/pkg/auth"}] = true
		g.imports[pbinfo.ImportSpec{Path: "github.com/microsoft/nodesdk/helper"}] = true
		g.imports[pbinfo.ImportSpec{Path: "k8s.io/klog"}] = true

	}

	// Connection()
	{
		p("// Connection returns the client's connection to the API service.")
		p("func (c *%sClient) Connection() *grpc.ClientConn {", servName)
		p("  return c.conn")
		p("}")
		p("")
	}

	// Close()
	{
		p("// Close closes the connection to the API service. The user should invoke this when")
		p("// the client is no longer required.")
		p("func (c *%sClient) Close() error {", servName)
		p("  return c.conn.Close()")
		p("}")
		p("")
	}
	
	return nil
}

func (g *generator) transformersInit() error {
	p := g.printf

	for kh,v := range g.sdkmapin {
		s := v.(map[interface{}]interface{})
		
		topMethod := fmt.Sprintf("get%s", kh)
		itemList := []ItemSdk{ ItemSdk{
			mapPart : s,
			name : kh,
			isRepeated: getMetaData(s["metadata"]).isRepeated,
			functionName: topMethod,
		}}
		stringPath := strings.Split(getMetaData(s["metadata"]).inSpec, ";")
		inSpec := stringPath[0]

		g.imports[pbinfo.ImportSpec{Name: stringPath[0], Path: stringPath[1]}] = true
		
		count := 0
		for  {
			if itemList[count].inSpec != "" {
				inSpec = itemList[count].inSpec
			}
			if itemList[count].isRepeated {
				itemList = append(itemList, g.generateTransFormerRepeated(itemList[count].mapPart, itemList[count].functionName, inSpec, itemList[count].name)...)
			} else if itemList[count].isEnum {
				g.generateTransFormerEnum(itemList[count].mapPart, itemList[count].functionName, inSpec, itemList[count].name)
			} else {
				itemList = append(itemList, g.generateTransFormer(itemList[count].mapPart, itemList[count].functionName, inSpec, itemList[count].name)...)
			}
			if count >= len(itemList) -1 {
				break
			}
			count++;
		}
	}

	for kh,v := range g.sdkmapout {
		s := v.(map[interface{}]interface{})
		
		sdkMapping := getMetaData(s["metadata"]).mapping
		topMethod := fmt.Sprintf("getSDK%s", kh)
		sdkNameSingle := strings.TrimSuffix(kh, "s")

		nextUp := s[kh].(map[interface{}]interface{})

		stringPath := strings.Split(getMetaData(s["metadata"]).inSpec, ";")
		inSpec := stringPath[0]

		p("func get%sFromResponse(response *%s.%s) []*%s {", kh, inSpec, sdkMapping, sdkNameSingle)
			p("return getSDK%s(response.%s)", kh, kh)
		p("}")
 
		itemList := []ItemSdk{ ItemSdk{
				mapPart : nextUp,
				name : kh,
				isRepeated: true,
				functionName: topMethod,
			}}
		count := 0
		for  {
			if itemList[count].inSpec != "" {
				inSpec = itemList[count].inSpec
			}
			if itemList[count].isRepeated {
				itemList = append(itemList, g.generateTransFormerRepeatedOut(itemList[count].mapPart, itemList[count].functionName, inSpec, itemList[count].name)...)
			} else {
				itemList = append(itemList, g.generateTransFormerOut(itemList[count].mapPart, itemList[count].functionName, inSpec, itemList[count].name)...)
			}
			if count >= len(itemList) -1 {
				break
			}
			count++;
		}
	}
	return nil
}

func (g *generator) clientInitStruct(servName string, messages []*descriptor.DescriptorProto, enums []*descriptor.EnumDescriptorProto, outDir string) error {
	p := g.printf

	for _, message := range messages {
		p("type %s struct {", message.GetName())
		for _, field := range message.GetField() {
			typ, _ := g.GoType(field, servName, outDir)
			p("	%s			%s", golangGen.CamelCase(field.GetName()), typ)
		}
		p("}")
		p("")
	}

	for _, enu := range enums {
		
		p("type %s int32", enu.GetName())
		p("const (")
		for i, v := range enu.GetValue() {
			p(" %s_%s %s = %v", enu.GetName(), v.GetName(), enu.GetName(), i)
		}
		p(")")
		p("")

		
	}

	return nil
}


// GoType returns a string representing the type name, and the wire type
func (g *generator) GoType(field *descriptor.FieldDescriptorProto, pkgName string, outDir string) (typ string, wire string) {
	// TODO: Options.
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		typ, wire = "float64", "fixed64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		typ, wire = "float32", "fixed32"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		typ, wire = "int64", "varint"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		typ, wire = "uint64", "varint"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		typ, wire = "int32", "varint"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		typ, wire = "uint32", "varint"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		typ, wire = "uint64", "fixed64"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		typ, wire = "uint32", "fixed32"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		typ, wire = "bool", "varint"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		typ, wire = "string", "bytes"
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		desc := g.ga.ObjectNamed(field.GetTypeName())
		typ, wire = "*"+g.ga.TypeName(desc), "group"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		desc := g.ga.ObjectNamed(field.GetTypeName())

		path, err := checkIfLocalPath(string(desc.GoImportPath()), pkgName, outDir)
		if err == nil {
			g.imports[pbinfo.ImportSpec{Path: path}] = true
		}
		typ, wire = "*"+ g.ga.TypeName(desc), "bytes"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		typ, wire = "[]byte", "bytes"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		desc := g.ga.ObjectNamed(field.GetTypeName())
		typ, wire = g.ga.TypeName(desc), "varint"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		typ, wire = "int32", "fixed32"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		typ, wire = "int64", "fixed64"
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		typ, wire = "int32", "zigzag32"
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		typ, wire = "int64", "zigzag64"
	default:
	}
	if isRepeated(field) {
		typ = "[]" + typ
	}

	return
}

// Is this field repeated?
func isRepeated(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

func checkIfLocalPath(path string, pkgName string, outDir string) (string, error) {
	if strings.Contains(path, "wssd") {
		test := strings.Split(path, "/")
		if test[len(test) - 1] == pkgName {
			return "", fmt.Errorf("Dupe")
		}
		return outDir + "/" + test[len(test) - 1], nil
	}
	return path, nil
}
