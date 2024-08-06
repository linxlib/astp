package types

import (
	"fmt"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/conv"
	"reflect"
	"strings"
)

type ElementType string

const (
	ElementNone   ElementType = "NONE"
	ElementStruct ElementType = "STRUCT"
	ElementField  ElementType = "FIELD"
	ElementConst  ElementType = "CONST"

	ElementVar       ElementType = "VAR"
	ElementMethod    ElementType = "METHOD"
	ElementFunc      ElementType = "FUNC"
	ElementEnum      ElementType = "ENUM"
	ElementInterface ElementType = "INTERFACE"
	ElementReceiver  ElementType = "RECEIVER"
	ElementGeneric   ElementType = "TYPE_PARAM"
	ElementParam     ElementType = "PARAM"
	ElementResult    ElementType = "RESULT"
	ElementConstrain ElementType = "CONSTRAIN"
)
const (
	PackageBuiltIn     = "builtin"
	PackageThisPackage = "this"
	PackageOther       = "other"
	PackageThird       = "third"
	PackageIgnore      = "ignore"
)

var builtInPackages = map[string]bool{
	"mime":    true,
	"time":    true,
	"errors":  true,
	"net":     true,
	"go":      true,
	"math":    true,
	"strconv": true,
	"path":    true,
	"os":      true,
}

func CheckPackage(modPkg string, pkg string) string {
	if pkg == PackageThisPackage || (strings.HasPrefix(pkg, modPkg) && pkg == modPkg) {
		return PackageThisPackage
	}
	idx := strings.Index(pkg, "/")
	var pkgPrefix string
	if idx < 0 {
		pkgPrefix = pkg
	} else {
		pkgPrefix = pkg[:idx]
	}
	if _, ok := builtInPackages[pkgPrefix]; ok {
		return PackageBuiltIn
	}
	if pkg == PackageBuiltIn {
		return PackageBuiltIn
	}
	if strings.HasPrefix(pkg, modPkg) {
		return PackageOther
	}
	return PackageThird
}

type Element struct {
	Name          string      `json:",omitempty"` //名称
	PackagePath   string      `json:",omitempty"` //包路径
	PackageName   string      `json:",omitempty"` //包名
	ElementType   ElementType `json:",omitempty"` //当前元素类型
	ItemType      ElementType `json:",omitempty"` //元素类型
	Item          *Element    `json:",omitempty"`
	Index         int
	Tag           reflect.StructTag `json:"-"`
	TypeString    string            `json:",omitempty"`
	ElementString string            `json:",omitempty"`
	Signature     string            `json:",omitempty"`
	Actual        *Element          `json:",omitempty"`
	Docs          []string          `json:",omitempty"`
	Comment       string            `json:",omitempty"`
	rType         reflect.Type      `json:"-"`
	rValue        reflect.Value     `json:"-"`
	Value         any
	FromParent    bool //表示当前元素 包含从父级继承而来的字段、方法、文档 或者 表示当前元素是继承而来
	Elements      map[ElementType][]*Element
}

func copySlice(src []*Element) []*Element {
	if src == nil {
		return nil
	}
	result := make([]*Element, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}

		result[i] = t.Clone()
	}
	return result
}

func (b *Element) String() string {
	return fmt.Sprintf("%s.%s", b.PackagePath, b.ElementString)
}
func (b *Element) Private() bool {
	return internal.IsPrivate(b.Name)
}
func (b *Element) Generic() bool {
	return (b.Elements[ElementGeneric] != nil && len(b.Elements[ElementGeneric]) > 0) || b.ItemType == ElementGeneric
}
func (b *Element) SetRType(t reflect.Type) {
	b.rType = t
}
func (b *Element) GetRType() reflect.Type {
	return b.rType
}
func (b *Element) SetRValue(value reflect.Value) {
	b.rValue = value
}
func (b *Element) GetRValue() reflect.Value {
	return b.rValue
}

func (b *Element) GetSignature() string {
	switch b.ElementType {
	case ElementStruct:
		builder := strings.Builder{}
		builder.WriteString(b.PackageName)
		builder.WriteRune('.')
		builder.WriteString(b.Name)
		if b.Generic() {
			builder.WriteRune('[')
			for i, typeParam := range b.Elements[ElementGeneric] {
				if i > 0 {
					builder.WriteRune(',')
				}
				builder.WriteString(typeParam.Name)
				builder.WriteRune(' ')
				builder.WriteString(typeParam.ElementString)
			}
			builder.WriteRune(']')
		}
		return builder.String()
	case ElementField:
		builder := strings.Builder{}
		builder.WriteString(b.Name)
		builder.WriteRune(' ')
		if b.ItemType == ElementStruct {
			builder.WriteString(b.Item.PackageName)
			builder.WriteRune('.')
			builder.WriteString(b.Item.Name)
		}

		if b.Generic() {
			builder.WriteRune('[')
			for i, typeParam := range b.Elements[ElementGeneric] {
				if i > 0 {
					builder.WriteRune(',')
				}
				builder.WriteString(typeParam.Name)
				builder.WriteRune(' ')
				builder.WriteString(typeParam.ElementString)
			}
			builder.WriteRune(']')
		}
		return builder.String()
	case ElementConst, ElementEnum:
		builder := strings.Builder{}
		builder.WriteString(b.Name)
		builder.WriteRune(' ')
		builder.WriteString(b.TypeString)
		builder.WriteRune('=')
		builder.WriteString(conv.String(b.Value))
		return builder.String()
	case ElementVar:
		builder := strings.Builder{}
		builder.WriteString(b.Name)
		builder.WriteRune(' ')
		builder.WriteString(b.TypeString)
		builder.WriteRune('=')
		builder.WriteString(conv.String(b.Value))
		return builder.String()
	case ElementMethod:
		builder := strings.Builder{}
		builder.WriteString("func (")
		for i, element := range b.Elements[ElementReceiver] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.GetSignature())
		}
		builder.WriteRune(')')
		builder.WriteString(b.Name)
		builder.WriteRune('(')
		for i, element := range b.Elements[ElementParam] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.GetSignature())
		}
		builder.WriteRune(')')
		builder.WriteRune('(')
		for i, element := range b.Elements[ElementResult] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.GetSignature())
		}
		builder.WriteRune(')')
		builder.WriteString("{}")
		return builder.String()

	case ElementFunc:
		builder := strings.Builder{}
		builder.WriteString("func ")
		builder.WriteString(b.Name)

		if b.Generic() {
			builder.WriteRune('[')
			for i, typeParam := range b.Elements[ElementGeneric] {
				if i > 0 {
					builder.WriteRune(',')
				}
				builder.WriteString(typeParam.GetSignature())
			}
			builder.WriteRune(']')
		}

		builder.WriteRune('(')
		for i, element := range b.Elements[ElementParam] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.GetSignature())
		}
		builder.WriteRune(')')
		builder.WriteRune('(')
		for i, element := range b.Elements[ElementResult] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.GetSignature())
		}
		builder.WriteRune(')')
		builder.WriteString("{}")
		return builder.String()
	case ElementInterface:
		return "a interface " + b.Name
	case ElementReceiver:
		builder := strings.Builder{}
		builder.WriteString(b.Name)
		builder.WriteRune(' ')
		builder.WriteString(b.TypeString)
		return builder.String()
	case ElementParam:
		builder := strings.Builder{}
		builder.WriteString(b.Name)
		builder.WriteRune(' ')
		builder.WriteString(b.TypeString)
		return builder.String()
	case ElementResult:
		builder := strings.Builder{}
		builder.WriteString(b.Name)
		builder.WriteRune(' ')
		builder.WriteString(b.TypeString)
		return builder.String()
	default:
		return b.String()
	}
}

func (b *Element) Clone(i ...int) *Element {
	if b == nil {
		return nil
	}
	newIndex := b.Index
	if len(i) > 0 {
		newIndex = i[0]
	}
	var e = &Element{
		Name:        b.Name,
		PackagePath: b.PackagePath,
		PackageName: b.PackageName,
		ElementType: b.ElementType,
		ItemType:    b.ItemType,
		Item:        b.Item.Clone(),
		Index:       newIndex,
		Tag:         b.Tag,
		TypeString:  b.TypeString,
		Actual:      b.Actual.Clone(),
		Docs:        b.Docs,
		Comment:     b.Comment,
		Value:       b.Value,
		Elements:    nil,
	}
	if b.Elements != nil {
		e.Elements = make(map[ElementType][]*Element)
		for elementType, elements := range b.Elements {
			e.Elements[elementType] = copySlice(elements)
		}
	}

	return e
}

// TODO: 需要判断 1. 内置类型 2. 内置包 3. 第三方包
func PackagePath(pkgName string, typeName string) string {
	if internal.IsInternalType(typeName) {
		return PackageBuiltIn
	}
	if pkgName == "" {
		return PackageThisPackage
	} else {

		return ""
	}

}
