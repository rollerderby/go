package main

import (
	"fmt"
	"io"
)

func buildStates(w io.Writer) {
	fmt.Fprintf(w, "var (\n")
	for _, tDef := range typeDefs {
		if tDef.root == "" {
			continue
		}
		fmt.Fprintf(w, "%v *%v\n", tDef.accessorName, tDef.accessorStruct)
	}
	fmt.Fprintf(w, ")\n")

	fmt.Fprintf(w, "\n\nfunc initializeState() error {\n")
	fmt.Fprintf(w, "\tstate.Root.Lock()\n")
	fmt.Fprintf(w, "\tdefer state.Root.Unlock()\n")
	for _, tDef := range typeDefs {
		if tDef.root == "" {
			continue
		}
		fmt.Fprintf(w, "\n%v = new%v(new%v().(*state.%v))\n", tDef.accessorName, tDef.accessorStruct, tDef.stateStruct, tDef.stateType)
		fmt.Fprintf(w, "if err := state.Root.Add(%#v, %#v, %v.state); err != nil { return err }\n", tDef.root, tDef.savePath, tDef.accessorName)
	}
	fmt.Fprintf(w, "return nil")
	fmt.Fprintf(w, "}")

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
			ct := findType(tDef.childType)
			return fmt.Sprintf("new%v", ct.stateStruct)
		}
	}

	for _, tDef := range typeDefs {
		fmt.Fprintf(w, "\n\nfunc new%v() state.Value {\n", tDef.stateStruct)
		switch tDef.stateType {
		case "Hash":
			ct := findType(tDef.childType)
			fmt.Fprintf(w, "return state.NewHashOf(new%v)()\n", ct.stateStruct)
		case "Object":
			fmt.Fprintf(w, "return &state.Object{\n")
			fmt.Fprintf(w, "	Definition: state.ObjectDef{\n")
			fmt.Fprintf(w, "		Name: %#v,\n", tDef.name)
			fmt.Fprintf(w, "		Values: []state.ObjectValueDef{\n")
			for _, fieldDef := range tDef.fields {
				switch fieldDef.stateType {
				case "Array":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewArrayOf(%v)},\n", fieldDef.name, childInit(fieldDef))
				case "Bool":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewBool},\n", fieldDef.name)
				case "Date":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewDate},\n", fieldDef.name)
				case "Enum":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewEnumOf(%#v...)},\n", fieldDef.name, fieldDef.enumValues)
				case "GUID":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewGUID},\n", fieldDef.name)
				case "Hash":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewHashOf(%v)},\n", fieldDef.name, childInit(fieldDef))
				case "Number":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewNumber},\n", fieldDef.name)
				case "String":
					fmt.Fprintf(w, "state.ObjectValueDef{Name: %#v, Initializer: state.NewString},\n", fieldDef.name)
				default:
					log.Errorf("Unknown StateType(%q) in %v.%v", fieldDef.stateType, tDef.name, fieldDef.name)
				}
			}
			fmt.Fprintf(w, "		},")
			fmt.Fprintf(w, "	},")
			fmt.Fprintf(w, "}")
		default:
			log.Errorf("Unknown StateType(%q) in %v", tDef.stateType, tDef.name)
			fmt.Fprintf(w, "return nil;\n")
		}
		fmt.Fprintf(w, "}")
	}
}
