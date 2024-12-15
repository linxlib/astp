package astp

import (
	"fmt"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/conv"
	"reflect"
	"strings"
)

type ElementType string

const (
	ElementNone ElementType = "NONE"

	ElementStruct      ElementType = "STRUCT"
	ElementArrayStruct ElementType = "ARRAY_STRUCT"
	ElementField       ElementType = "FIELD"
	ElementConst       ElementType = "CONST"
	ElementVar         ElementType = "VAR"
	ElementFunc        ElementType = "FUNC"
	ElementInterface   ElementType = "INTERFACE"

	ElementMethod    ElementType = "METHOD"
	ElementEnum      ElementType = "ENUM"
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

// ParseIt 返回一个类型是否需要进行解析（第三方包 系统内置包）
func ParseIt(modPkg string, pkgPath string) bool {
	return true
}

// CheckPackage 返回某个包是何种类型的包
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
	Name        string      `json:",omitempty" yaml:",omitempty"`
	PackagePath string      `json:",omitempty" yaml:",omitempty"` //包路径
	PackageName string      `json:",omitempty" yaml:",omitempty"` //包名
	ElementType ElementType `json:",omitempty" yaml:",omitempty"` //当前元素类型

	ItemType      ElementType                `json:",omitempty" yaml:",omitempty"` //Item的元素类型
	Item          *Element                   `json:",omitempty" yaml:",omitempty"`
	IsItemSlice   bool                       `json:",omitempty" yaml:",omitempty"`
	IsEnumString  bool                       `json:",omitempty" yaml:",omitempty"`
	Index         int                        `json:",omitempty" yaml:",omitempty"`
	TagString     string                     `json:",omitempty" yaml:",omitempty"`
	TypeString    string                     `json:",omitempty" yaml:",omitempty"` //字段类型名
	ElementString string                     `json:",omitempty" yaml:",omitempty"`
	Docs          []string                   `json:",omitempty" yaml:",omitempty"` //上方的文档
	Comment       string                     `json:",omitempty" yaml:",omitempty"` //后方的注释
	rType         reflect.Type               `json:"-"`
	rValue        reflect.Value              `json:"-"`
	Value         any                        `json:",omitempty" yaml:",omitempty"` //值 一般为枚举时用到
	FromParent    bool                       `json:",omitempty" yaml:",omitempty"` //表示当前元素 包含从父级继承而来的字段、方法、文档 或者 表示当前元素是继承而来
	Elements      map[ElementType][]*Element `json:",omitempty" yaml:",omitempty"` // 子成员 比如字段 方法 泛型类型
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
func (b *Element) SetValue(i any) {
	b.Value = i
}
func (b *Element) GetValue() any {
	return b.Value
}

func (b *Element) Signature() string {
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
			builder.WriteString(element.Signature())
		}
		builder.WriteRune(')')
		builder.WriteString(b.Name)
		builder.WriteRune('(')
		for i, element := range b.Elements[ElementParam] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.Signature())
		}
		builder.WriteRune(')')
		builder.WriteRune('(')
		for i, element := range b.Elements[ElementResult] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.Signature())
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
				builder.WriteString(typeParam.Signature())
			}
			builder.WriteRune(']')
		}

		builder.WriteRune('(')
		for i, element := range b.Elements[ElementParam] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.Signature())
		}
		builder.WriteRune(')')
		builder.WriteRune('(')
		for i, element := range b.Elements[ElementResult] {
			if i > 0 {
				builder.WriteRune(',')
			}
			builder.WriteString(element.Signature())
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
		Name:         b.Name,
		PackagePath:  b.PackagePath,
		PackageName:  b.PackageName,
		ElementType:  b.ElementType,
		ItemType:     b.ItemType,
		Item:         b.Item.Clone(),
		Index:        newIndex,
		TypeString:   b.TypeString,
		Docs:         b.Docs,
		Comment:      b.Comment,
		IsEnumString: b.IsEnumString,
		TagString:    b.TagString,
		Value:        b.Value,
		Elements:     nil,
	}
	if b.Elements != nil {
		e.Elements = make(map[ElementType][]*Element)
		for elementType, elements := range b.Elements {
			e.Elements[elementType] = copySlice(elements)
		}
	}

	return e
}

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

func (b *Element) VisitElements(elementType ElementType, check func(element *Element) bool, f func(element *Element)) {
	for _, e := range b.Elements[elementType] {
		if !check(e) {
			continue
		}
		f(e)
	}
}
func (b *Element) VisitElementsAll(elementType ElementType, f func(element *Element)) {
	for _, e := range b.Elements[elementType] {
		f(e)
	}
}
func (b *Element) ElementsAll(elementType ElementType) []*Element {
	return b.Elements[elementType]
}
func (b *Element) MustGetElement(elementType ElementType) *Element {
	elements := b.Elements[elementType]
	for _, element := range elements {
		return element
	}
	return nil
}

func (b *Element) GetTag() reflect.StructTag {
	return reflect.StructTag(strings.Trim(b.TagString, "`"))
}
