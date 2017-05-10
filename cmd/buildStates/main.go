package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
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
		types := extractTypes(jValue, "")

		log.Debugf("Generating state code in %v for %v", *dir, *packageName)
		buildState(*packageName, path.Join(*dir, "state.go"), types)
		buildHelpers(*packageName, path.Join(*dir, "state_helpers.go"), types)
	}
}

type typeDef struct {
	name              string
	root              string
	savePath          string
	stateType         string
	childType         string
	initFunc          bool
	fields            []*typeDef
	enumValues        []string
	specialHelperName string
}

func extractTypes(jValue json.Array, helperPrefix string) []*typeDef {
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

			if helperPrefix != "" {
				switch tDef.stateType {
				case "Array":
					tDef.specialHelperName = helperPrefix + tDef.name
				case "Hash":
					tDef.specialHelperName = helperPrefix + tDef.name
				case "Object":
					tDef.specialHelperName = helperPrefix + tDef.name
				}
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

func buildState(packageName, goFile string, typeDefs []*typeDef) error {
	buf := &bytes.Buffer{}

	log.Debugf("Building state file to %q", goFile)

	fmt.Fprintf(buf, "package %v\n\nimport \"github.com/rollerderby/go/state\"\n\n", packageName)
	fmt.Fprintf(buf, "// Auto generated using buildStates command\n\n")

	fmt.Fprintf(buf, "var (\n")
	for _, tDef := range typeDefs {
		if tDef.root == "" {
			continue
		}
		fmt.Fprintf(buf, "%v *%vHelper\n", tDef.name, tDef.name)
	}
	fmt.Fprintf(buf, ")\n")

	fmt.Fprintf(buf, "\n\nfunc initializeState() error {\n")
	fmt.Fprintf(buf, "\tstate.Root.Lock()\n")
	fmt.Fprintf(buf, "\tdefer state.Root.Unlock()\n")
	for _, tDef := range typeDefs {
		if tDef.root == "" {
			continue
		}
		fmt.Fprintf(buf, "\n%v = new%vHelper(new%v().(*state.%v))\n", tDef.name, tDef.name, tDef.name, tDef.stateType)
		fmt.Fprintf(buf, "if err := state.Root.Add(%#v, %#v, %v.state); err != nil { return err }\n", tDef.root, tDef.savePath, tDef.name)
	}
	fmt.Fprintf(buf, "return nil")
	fmt.Fprintf(buf, "}")

	childInit := func(tDef *typeDef) string {
		switch tDef.childType {
		case "Bool":
			return "state.NewBool"
		case "Date":
			return "state.NewDate"
		case "Enum":
			return fmt.Sprintf("state.NewEnumOf(%#v...)", tDef.enumValues)
		case "GUID":
			return "state.NewGUID"
		case "Number":
			return "state.NewNumber"
		case "String":
			return "state.NewString"
		default:
			return fmt.Sprintf("new%v", tDef.childType)
		}
	}

	for _, tDef := range typeDefs {
		fmt.Fprintf(buf, "\n\nfunc new%v() state.Value {\n", tDef.name)
		switch tDef.stateType {
		case "Hash":
			fmt.Fprintf(buf, "return state.NewHashOf(new%v)()\n", tDef.childType)
		case "Object":
			fmt.Fprintf(buf, "return &state.Object{\n")
			fmt.Fprintf(buf, "	Definition: state.ObjectDef{\n")
			fmt.Fprintf(buf, "		Name: %#v,\n", tDef.name)
			fmt.Fprintf(buf, "		Values: []state.ObjectValueDef{\n")
			for _, fieldDef := range tDef.fields {
				switch fieldDef.stateType {
				case "Array":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewArrayOf(%v)},\n", fieldDef.name, childInit(fieldDef))
				case "Bool":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewBool},\n", fieldDef.name)
				case "Date":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewDate},\n", fieldDef.name)
				case "Enum":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewEnumOf(%#v...)},\n", fieldDef.name, fieldDef.enumValues)
				case "GUID":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewGUID},\n", fieldDef.name)
				case "Hash":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewHashOf(%v)},\n", fieldDef.name, childInit(fieldDef))
				case "Number":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewNumber},\n", fieldDef.name)
				case "String":
					fmt.Fprintf(buf, "state.ObjectValueDef{Name: %#v, Initializer: state.NewString},\n", fieldDef.name)
				default:
					log.Errorf("Unknown StateType(%q) in %v.%v", fieldDef.stateType, tDef.name, fieldDef.name)
				}
			}
			fmt.Fprintf(buf, "		},")
			fmt.Fprintf(buf, "	},")
			fmt.Fprintf(buf, "}")
		default:
			log.Errorf("Unknown StateType(%q) in %v", tDef.stateType, tDef.name)
			fmt.Fprintf(buf, "return nil;\n")
		}
		fmt.Fprintf(buf, "}")
	}

	if err := saveGoCode(goFile, buf.Bytes()); err != nil {
		log.Errorf("Error saving go code: %v", err)
		return err
	}

	return nil
}

func buildHelper(w io.Writer, helperName string, tDef *typeDef, typeDefs []*typeDef) {
	var stateType = "state.Value"
	switch tDef.stateType {
	case "Array":
		stateType = "*state.Array"
	case "Hash":
		stateType = "*state.Hash"
	case "Object":
		stateType = "*state.Object"
	default:
		log.Errorf("Unhandled StateType(%q) in %v", tDef.stateType, tDef.name)
	}
	fmt.Fprintf(w, "\n\ntype %vHelper struct {\n", helperName)
	fmt.Fprintf(w, "	state %v", stateType)
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func new%vHelper(state %v) *%vHelper {\n", helperName, stateType, helperName)
	fmt.Fprintf(w, "	ret := &%vHelper{state}\n", helperName)
	if tDef.initFunc {
		fmt.Fprintf(w, "	ret.init()\n")
	}
	fmt.Fprintf(w, "	return ret\n")
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (h *%vHelper) Path() string {\n", helperName)
	fmt.Fprintf(w, "	return h.state.Path()\n")
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (h *%vHelper) JSON(indent bool) string {\n", helperName)
	fmt.Fprintf(w, "	return h.state.JSON(false).JSON(indent)\n")
	fmt.Fprintf(w, "}\n\n")

	switch tDef.stateType {
	case "Array":
		helperArray(w, helperName, tDef, typeDefs)
	case "Hash":
		helperHash(w, helperName, tDef, typeDefs)
	case "Object":
		helperObject(w, helperName, tDef, typeDefs)
	default:
		log.Errorf("Unhandled StateType(%q) in %v", tDef.stateType, tDef.name)
		fmt.Fprintf(w, "state state.Value\n")
	}

	for _, f := range tDef.fields {
		if f.specialHelperName != "" {
			buildHelper(w, f.specialHelperName, f, typeDefs)
		}
	}
}

func buildHelpers(packageName, goFile string, typeDefs []*typeDef) error {
	buf := &bytes.Buffer{}

	log.Debugf("Building helper file to %q", goFile)

	fmt.Fprintf(buf, "package %v\n\nimport \"github.com/rollerderby/go/state\"\n\n", packageName)
	fmt.Fprintf(buf, "// Auto generated using buildStates command\n\n")

	for _, tDef := range typeDefs {
		buildHelper(buf, tDef.name, tDef, typeDefs)
	}

	if err := saveGoCode(goFile, buf.Bytes()); err != nil {
		log.Errorf("Error saving go code: %v", err)
		return err
	}

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

func helperArray(w io.Writer, helperName string, tDef *typeDef, typeDefs []*typeDef) {
	fmt.Fprintf(w, "func (h *%vHelper) Clear() {", helperName)
	fmt.Fprintf(w, "	h.state.Clear()\n")
	fmt.Fprintf(w, "}\n\n")

	if !isSimpleType(tDef.childType) {
		childStateType := "state.Value"
		var childDef *typeDef
		for _, c := range typeDefs {
			if c.name == tDef.childType {
				childStateType = "*state." + c.stateType
				childDef = c
				break
			}
		}

		fmt.Fprintf(w, "func (h *%vHelper) New() (*%vHelper, error) {", helperName, tDef.childType)
		fmt.Fprintf(w, "	stateObj, err := h.state.NewEmptyElement()\n")
		fmt.Fprintf(w, "	if err != nil {\n")
		fmt.Fprintf(w, "		return nil, err\n")
		fmt.Fprintf(w, "	}\n")
		fmt.Fprintf(w, "	return new%vHelper(stateObj.(%v)), nil\n", tDef.childType, childStateType)
		fmt.Fprintf(w, "}\n\n")
		fmt.Fprintf(w, "func (h *%vHelper) Values() []*%vHelper {", helperName, tDef.childType)
		fmt.Fprintf(w, "	var ret []*%vHelper\n", tDef.childType)
		fmt.Fprintf(w, "	for _, val := range h.state.Values() {\n")
		fmt.Fprintf(w, "		ret = append(ret, new%vHelper(val.(%v)))\n", tDef.childType, childStateType)
		fmt.Fprintf(w, "	}\n")
		fmt.Fprintf(w, "	return ret\n")
		fmt.Fprintf(w, "}\n\n")
		if childDef != nil && childDef.stateType == "Object" {
			for _, field := range childDef.fields {
				goType := goType(field.stateType)
				if goType != "" {
					fmt.Fprintf(w, "func (h *%vHelper) FindBy%v(lookFor %v) []*%vHelper {\n", helperName, field.name, goType, tDef.childType)
					fmt.Fprintf(w, "	var ret []*%vHelper\n", tDef.childType)
					fmt.Fprintf(w, "	for _, obj := range h.Values() {\n")
					fmt.Fprintf(w, "		if obj.%v() == lookFor {\n", field.name)
					fmt.Fprintf(w, "			ret = append(ret, obj)\n")
					fmt.Fprintf(w, "		}\n")
					fmt.Fprintf(w, "	}\n")
					fmt.Fprintf(w, "	return ret\n")
					fmt.Fprintf(w, "}\n\n")
				}
			}
		}
	} else {
		goType := goType(tDef.childType)

		fmt.Fprintf(w, "func (h *%vHelper) Add(v %v) error {", helperName, goType)
		fmt.Fprintf(w, "	elem, err := h.state.NewEmptyElement()\n")
		fmt.Fprintf(w, "	if err != nil { return err }\n")
		fmt.Fprintf(w, "	return elem.(*state.%v).SetValue(v)\n", tDef.childType)
		fmt.Fprintf(w, "}\n\n")
		fmt.Fprintf(w, "func (h *%vHelper) New() (*state.%v, error) {", helperName, tDef.childType)
		fmt.Fprintf(w, "	elem, err := h.state.NewEmptyElement()\n")
		fmt.Fprintf(w, "	if err != nil { return nil, err }\n")
		fmt.Fprintf(w, "	return elem.(*state.%v), nil\n", tDef.childType)
		fmt.Fprintf(w, "}\n\n")
		fmt.Fprintf(w, "func (h *%vHelper) Values() []%v {\n", helperName, goType)
		fmt.Fprintf(w, "	var ret []%v\n", goType)
		fmt.Fprintf(w, "	for _, val := range h.state.Values() {\n")
		fmt.Fprintf(w, "		ret = append(ret, val.(*state.%v).Value())\n", tDef.childType)
		fmt.Fprintf(w, "	}\n")
		fmt.Fprintf(w, "	return ret\n")
		fmt.Fprintf(w, "}\n\n")
	}
}

func helperHash(w io.Writer, helperName string, tDef *typeDef, typeDefs []*typeDef) {
	childStateType := "state.Value"
	var childDef *typeDef
	for _, c := range typeDefs {
		if c.name == tDef.childType {
			childStateType = "*state." + c.stateType
			childDef = c
			break
		}
	}
	fmt.Fprintf(w, "func (h *%vHelper) New(key string) (*%vHelper, error) {", helperName, tDef.childType)
	fmt.Fprintf(w, "	stateObj, err := h.state.NewEmptyElement(key)\n")
	fmt.Fprintf(w, "	if err != nil {\n")
	fmt.Fprintf(w, "		return nil, err\n")
	fmt.Fprintf(w, "	}\n")
	fmt.Fprintf(w, "	return new%vHelper(stateObj.(%v)), nil\n", tDef.childType, childStateType)
	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "func (h *%vHelper) Values() []*%vHelper {", helperName, tDef.childType)
	fmt.Fprintf(w, "	var ret []*%vHelper\n", tDef.childType)
	fmt.Fprintf(w, "	for _, val := range h.state.Values() {\n")
	fmt.Fprintf(w, "		ret = append(ret, new%vHelper(val.(%v)))\n", tDef.childType, childStateType)
	fmt.Fprintf(w, "	}\n")
	fmt.Fprintf(w, "	return ret\n")
	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "func (h *%vHelper) Keys() []string {", helperName)
	fmt.Fprintf(w, "	return h.state.Keys()\n")
	fmt.Fprintf(w, "}\n\n")

	if childDef != nil && childDef.stateType == "Object" {
		for idx, field := range childDef.fields {
			goType := goType(field.stateType)
			if goType != "" {
				if idx == 0 && field.name == "ID" && field.stateType == "GUID" {
					fmt.Fprintf(w, "func (h *%vHelper) Get(id %v) *%vHelper {\n", helperName, goType, tDef.childType)
					fmt.Fprintf(w, "	for _, obj := range h.Values() {\n")
					fmt.Fprintf(w, "		if obj.%v() == id {\n", field.name)
					fmt.Fprintf(w, "			return obj\n")
					fmt.Fprintf(w, "		}\n")
					fmt.Fprintf(w, "	}\n")
					fmt.Fprintf(w, "	return nil\n")
					fmt.Fprintf(w, "}\n\n")
				} else {
					fmt.Fprintf(w, "func (h *%vHelper) FindBy%v(lookFor %v) []*%vHelper {\n", helperName, field.name, goType, tDef.childType)
					fmt.Fprintf(w, "	var ret []*%vHelper\n", tDef.childType)
					fmt.Fprintf(w, "	for _, obj := range h.Values() {\n")
					fmt.Fprintf(w, "		if obj.%v() == lookFor {\n", field.name)
					fmt.Fprintf(w, "			ret = append(ret, obj)\n")
					fmt.Fprintf(w, "		}\n")
					fmt.Fprintf(w, "	}\n")
					fmt.Fprintf(w, "	return ret\n")
					fmt.Fprintf(w, "}\n\n")
				}
			}
		}
	}
}

func helperObject(w io.Writer, helperName string, tDef *typeDef, typeDefs []*typeDef) {
	for _, field := range tDef.fields {
		goType := goType(field.stateType)
		if goType != "" {
			fmt.Fprintf(w, "func (h *%vHelper) %v() %v {", helperName, field.name, goType)
			fmt.Fprintf(w, "	return h.state.Get(%#v).(*state.%v).Value()\n", field.name, field.stateType)
			fmt.Fprintf(w, "}\n\n")
			fmt.Fprintf(w, "func (h *%vHelper) Set%v(val %v) error {", helperName, field.name, goType)
			fmt.Fprintf(w, "	return h.state.Get(%#v).(*state.%v).SetValue(val)\n", field.name, field.stateType)
			fmt.Fprintf(w, "}\n\n")
		} else {
			switch field.stateType {
			case "Array":
				fmt.Fprintf(w, "func (h *%vHelper) %v() *%vHelper {", helperName, field.name, field.specialHelperName)
				fmt.Fprintf(w, "	return new%vHelper(h.state.Get(%#v).(*state.Array))\n", field.specialHelperName, field.name)
				fmt.Fprintf(w, "}\n")
			case "Hash":
				fmt.Fprintf(w, "func (h *%vHelper) %v() *%vHelper {", helperName, field.name, field.specialHelperName)
				fmt.Fprintf(w, "	return new%vHelper(h.state.Get(%#v).(*state.Hash))\n", field.specialHelperName, field.name)
				fmt.Fprintf(w, "}\n")
			case "Object":
				fmt.Fprintf(w, "func (h *%vHelper) %v() *%vHelper {", helperName, field.name, field.specialHelperName)
				fmt.Fprintf(w, "	return new%vHelper(h.state.Get(%#v).(*state.Object))\n", field.specialHelperName, field.name)
				fmt.Fprintf(w, "}\n")
			default:
				log.Errorf("Unhandled type: %v", field.stateType)
			}
		}
	}
}
