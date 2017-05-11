package main

import (
	"fmt"
	"io"
)

func buildAccessors(w io.Writer) {
	for _, tDef := range typeDefs {
		buildAccessor(w, tDef)
	}
}

func buildAccessor(w io.Writer, tDef *typeDef) {
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
	fmt.Fprintf(w, "\n\ntype %v struct {\n", tDef.accessorStruct)
	fmt.Fprintf(w, "	state %v", stateType)
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func new%v(state %v) *%v{\n", tDef.accessorStruct, stateType, tDef.accessorStruct)
	fmt.Fprintf(w, "	ret := &%v{state}\n", tDef.accessorStruct)
	if tDef.initFunc {
		fmt.Fprintf(w, "	ret.init()\n")
	}
	fmt.Fprintf(w, "	return ret\n")
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (h *%v) Path() string {\n", tDef.accessorStruct)
	fmt.Fprintf(w, "	return h.state.Path()\n")
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "func (h *%v) JSON(indent bool) string {\n", tDef.accessorStruct)
	fmt.Fprintf(w, "	return h.state.JSON(false).JSON(indent)\n")
	fmt.Fprintf(w, "}\n\n")

	switch tDef.stateType {
	case "Array":
		arrayAccessor(w, tDef)
	case "Hash":
		hashAccessor(w, tDef)
	case "Object":
		objectAccessor(w, tDef)
	default:
		log.Errorf("Unhandled StateType(%q) in %v", tDef.stateType, tDef.name)
		fmt.Fprintf(w, "state state.Value\n")
	}

	for _, f := range tDef.fields {
		if f.accessorName != "" {
			buildAccessor(w, f)
		}
	}
}

func arrayAccessor(w io.Writer, tDef *typeDef) {
	fmt.Fprintf(w, "func (h *%v) Clear() {", tDef.accessorStruct)
	fmt.Fprintf(w, "	h.state.Clear()\n")
	fmt.Fprintf(w, "}\n\n")

	if !isSimpleType(tDef.childType) {
		childStateType := "state.Value"
		var childDef *typeDef
		if childDef = findType(tDef.childType); childDef != nil {
			childStateType = "*state." + childDef.stateType
		}

		fmt.Fprintf(w, "func (h *%v) New() (*%v, error) {", tDef.accessorStruct, tDef.childType)
		fmt.Fprintf(w, "	stateObj, err := h.state.NewEmptyElement()\n")
		fmt.Fprintf(w, "	if err != nil {\n")
		fmt.Fprintf(w, "		return nil, err\n")
		fmt.Fprintf(w, "	}\n")
		fmt.Fprintf(w, "	return new%v(stateObj.(%v)), nil\n", tDef.childType, childStateType)
		fmt.Fprintf(w, "}\n\n")
		fmt.Fprintf(w, "func (h *%v) Values() []*%v{", tDef.accessorStruct, tDef.childType)
		fmt.Fprintf(w, "	var ret []*%v\n", tDef.childType)
		fmt.Fprintf(w, "	for _, val := range h.state.Values() {\n")
		fmt.Fprintf(w, "		ret = append(ret, new%v(val.(%v)))\n", tDef.childType, childStateType)
		fmt.Fprintf(w, "	}\n")
		fmt.Fprintf(w, "	return ret\n")
		fmt.Fprintf(w, "}\n\n")
		if childDef != nil && childDef.stateType == "Object" {
			for _, field := range childDef.fields {
				goType := goType(field.stateType)
				if goType != "" {
					fmt.Fprintf(w, "func (h *%v) FindBy%v(lookFor %v) []*%v {\n", tDef.accessorStruct, field.name, goType, tDef.childType)
					fmt.Fprintf(w, "	var ret []*%v\n", tDef.childType)
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

		fmt.Fprintf(w, "func (h *%v) Add(v %v) error {", tDef.accessorStruct, goType)
		fmt.Fprintf(w, "	elem, err := h.state.NewEmptyElement()\n")
		fmt.Fprintf(w, "	if err != nil { return err }\n")
		fmt.Fprintf(w, "	return elem.(*state.%v).SetValue(v)\n", tDef.childType)
		fmt.Fprintf(w, "}\n\n")
		fmt.Fprintf(w, "func (h *%v) New() (*state.%v, error) {", tDef.accessorStruct, tDef.childType)
		fmt.Fprintf(w, "	elem, err := h.state.NewEmptyElement()\n")
		fmt.Fprintf(w, "	if err != nil { return nil, err }\n")
		fmt.Fprintf(w, "	return elem.(*state.%v), nil\n", tDef.childType)
		fmt.Fprintf(w, "}\n\n")
		fmt.Fprintf(w, "func (h *%v) Values() []%v {\n", tDef.accessorStruct, goType)
		fmt.Fprintf(w, "	var ret []%v\n", goType)
		fmt.Fprintf(w, "	for _, val := range h.state.Values() {\n")
		fmt.Fprintf(w, "		ret = append(ret, val.(*state.%v).Value())\n", tDef.childType)
		fmt.Fprintf(w, "	}\n")
		fmt.Fprintf(w, "	return ret\n")
		fmt.Fprintf(w, "}\n\n")
	}
}

func hashAccessor(w io.Writer, tDef *typeDef) {
	childStateType := "state.Value"
	var childDef *typeDef
	if childDef = findType(tDef.childType); childDef != nil {
		childStateType = "*state." + childDef.stateType
	}

	fmt.Fprintf(w, "func (h *%v) New(key string) (*%v, error) {", tDef.accessorStruct, tDef.childType)
	fmt.Fprintf(w, "	stateObj, err := h.state.NewEmptyElement(key)\n")
	fmt.Fprintf(w, "	if err != nil {\n")
	fmt.Fprintf(w, "		return nil, err\n")
	fmt.Fprintf(w, "	}\n")
	fmt.Fprintf(w, "	return new%v(stateObj.(%v)), nil\n", tDef.childType, childStateType)
	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "func (h *%v) Values() []*%v{", tDef.accessorStruct, tDef.childType)
	fmt.Fprintf(w, "	var ret []*%v\n", tDef.childType)
	fmt.Fprintf(w, "	for _, val := range h.state.Values() {\n")
	fmt.Fprintf(w, "		ret = append(ret, new%v(val.(%v)))\n", tDef.childType, childStateType)
	fmt.Fprintf(w, "	}\n")
	fmt.Fprintf(w, "	return ret\n")
	fmt.Fprintf(w, "}\n\n")
	fmt.Fprintf(w, "func (h *%v) Keys() []string {", tDef.accessorStruct)
	fmt.Fprintf(w, "	return h.state.Keys()\n")
	fmt.Fprintf(w, "}\n\n")

	if childDef != nil && childDef.stateType == "Object" {
		for idx, field := range childDef.fields {
			goType := goType(field.stateType)
			if goType != "" {
				if idx == 0 && field.name == "ID" && field.stateType == "GUID" {
					fmt.Fprintf(w, "func (h *%v) Get(id %v) *%v{\n", tDef.accessorStruct, goType, tDef.childType)
					fmt.Fprintf(w, "	for _, obj := range h.Values() {\n")
					fmt.Fprintf(w, "		if obj.%v() == id {\n", field.name)
					fmt.Fprintf(w, "			return obj\n")
					fmt.Fprintf(w, "		}\n")
					fmt.Fprintf(w, "	}\n")
					fmt.Fprintf(w, "	return nil\n")
					fmt.Fprintf(w, "}\n\n")
				} else {
					fmt.Fprintf(w, "func (h *%v) FindBy%v(lookFor %v) []*%v{\n", tDef.accessorStruct, field.name, goType, tDef.childType)
					fmt.Fprintf(w, "	var ret []*%v\n", tDef.childType)
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

func objectAccessor(w io.Writer, tDef *typeDef) {
	for _, field := range tDef.fields {
		goType := goType(field.stateType)
		if goType != "" {
			fmt.Fprintf(w, "func (h *%v) %v() %v {", tDef.accessorStruct, field.name, goType)
			fmt.Fprintf(w, "	return h.state.Get(%#v).(*state.%v).Value()\n", field.name, field.stateType)
			fmt.Fprintf(w, "}\n\n")
			fmt.Fprintf(w, "func (h *%v) Set%v(val %v) error {", tDef.accessorStruct, field.name, goType)
			fmt.Fprintf(w, "	return h.state.Get(%#v).(*state.%v).SetValue(val)\n", field.name, field.stateType)
			fmt.Fprintf(w, "}\n\n")
		} else {
			if field.accessorName != "" {
				fmt.Fprintf(w, "func (h *%v) %v() *%v{", tDef.accessorStruct, field.name, field.accessorName)
				fmt.Fprintf(w, "	return new%v(h.state.Get(%#v).(*state.%v))\n", field.accessorName, field.name, field.stateType)
				fmt.Fprintf(w, "}\n")
			} else {
				log.Errorf("Unhandled type: %v", field.stateType)
			}
		}
	}
}
