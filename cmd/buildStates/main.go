package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/rollerderby/go/json"
	"github.com/rollerderby/go/logger"
)

var log = logger.New("buildStates")

func main() {
	tmpDir, err := os.Getwd()
	if err != nil {
		tmpDir = "."
	}
	tmpPackage := os.Getenv("GOPACKAGE")
	if tmpPackage == "" {
		tmpPackage = "main"
	}
	tmpVerbose := os.Getenv("VERBOSE") == "1"

	dir := flag.String("dir", tmpDir, "Path to folder to generate state objects")
	packageName := flag.String("package", tmpPackage, "Package")
	verbose := flag.Bool("v", tmpVerbose, "Verbose")
	flag.Parse()

	if *verbose {
		log.SetLevel(logger.DEBUG)
	}

	data, err := ioutil.ReadFile(path.Join(*dir, "stateDef.json"))
	if err != nil {
		log.Fatalf("Cannot read state file: %v", err)
		os.Exit(1)
	}

	jValue, err := json.Decode(data)
	if err != nil {
		log.Fatalf("Cannot decode json: %v", err)
		os.Exit(1)
	}

	if jValue, ok := jValue.(json.Array); !ok {
		log.Fatal("Invalid json format")
		os.Exit(1)
	} else {
		typeDefs = extractTypes(jValue, "")

		log.Debugf("Generating state code in %v for %v", *dir, *packageName)
		buf := &bytes.Buffer{}

		goFile := path.Join(*dir, "state.go")
		log.Debugf("Building state file to %q", goFile)

		fmt.Fprintf(buf, "package %v\n\nimport \"github.com/rollerderby/go/state\"\n\n", *packageName)
		fmt.Fprintf(buf, "// Auto generated using buildStates command\n\n")

		buildStates(buf)
		buildAccessors(buf)

		if err := saveGoCode(goFile, buf.Bytes()); err != nil {
			log.Errorf("Error saving go code: %v", err)
		}
	}
}

type typeDef struct {
	name           string
	root           string
	savePath       string
	stateType      string
	childType      string
	initFunc       bool
	fields         []*typeDef
	enumValues     []string
	stateStruct    string
	accessorName   string
	accessorStruct string
}

var typeDefs []*typeDef

func extractTypes(jValue json.Array, accessorPrefix string) []*typeDef {
	getString := func(obj json.Object, field string) string {
		val, ok := obj[field].(*json.String)
		if !ok {
			return ""
		}
		delete(obj, field)
		return val.Get()
	}

	var ret []*typeDef
	for _, val := range jValue {
		if val, ok := val.(json.Object); ok {
			tDef := &typeDef{}

			tDef.name = getString(val, "Name")
			tDef.root = getString(val, "Root")
			tDef.savePath = getString(val, "SavePath")
			tDef.stateType = getString(val, "StateType")
			tDef.childType = getString(val, "ChildType")

			switch tDef.stateType {
			case "Array":
				tDef.stateStruct = "_state_" + tDef.name
				tDef.accessorName = accessorPrefix + tDef.name
				tDef.accessorStruct = accessorPrefix + tDef.name
			case "Hash":
				tDef.stateStruct = "_state_" + tDef.name
				tDef.accessorName = accessorPrefix + tDef.name
				tDef.accessorStruct = accessorPrefix + tDef.name
			case "Object":
				tDef.stateStruct = "_state_" + tDef.name
				tDef.accessorName = accessorPrefix + tDef.name
				tDef.accessorStruct = accessorPrefix + tDef.name
			}

			if tDef.root != "" {
				tDef.accessorStruct = "Root" + tDef.accessorStruct
			}

			if initFunc, ok := val["InitFunc"]; ok {
				tDef.initFunc = initFunc == json.True
				delete(val, "InitFunc")
			}

			if enumValues, ok := val["EnumValues"].(json.Array); ok {
				for _, enumValue := range enumValues {
					if enumValue, ok := enumValue.(*json.String); ok {
						tDef.enumValues = append(tDef.enumValues, enumValue.Get())
					}
				}
				delete(val, "EnumValues")
			}

			if fields, ok := val["Fields"].(json.Array); ok {
				tDef.fields = extractTypes(fields, tDef.name+"_")
				delete(val, "Fields")
			}

			if tDef.root != "" && tDef.savePath == "" {
				tDef.savePath = strings.ToLower(tDef.root)
			}

			if len(val) > 0 {
				log.Errorf("Unhandled fields: %v", val)
			}

			ret = append(ret, tDef)
		}
	}
	return ret
}

func saveGoCode(filename string, data []byte) error {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	in.Write(data)

	cmd := exec.Command("gofmt")
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ioutil.WriteFile(filename, data, 0644)
		return err
	}

	ioutil.WriteFile(filename, out.Bytes(), 0644)
	return nil
}

func goType(st string) string {
	switch st {
	case "Bool":
		return "bool"
	case "Date":
		return "string"
	case "Enum":
		return "string"
	case "GUID":
		return "string"
	case "Number":
		return "int64"
	case "String":
		return "string"
	}
	return ""
}

func isSimpleType(t string) bool {
	return goType(t) != ""
}

func findType(name string) *typeDef {
	for _, tDef := range typeDefs {
		if name == tDef.name {
			return tDef
		}
	}
	return nil
}
