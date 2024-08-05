package types

import (
	"fmt"
	"github.com/linxlib/astp/internal"
	"reflect"
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
	PackageThird       = "third"
	PackageIgnore      = "ignore"
)

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

func copySlice[T any](src []*T) []*T {
	if src == nil {
		return nil
	}
	result := make([]*T, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		v := *t
		result[i] = &v
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
	return b.Elements[ElementGeneric] != nil && len(b.Elements[ElementGeneric]) > 0
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
		Name:          b.Name,
		PackagePath:   b.PackagePath,
		PackageName:   b.PackageName,
		ElementType:   b.ElementType,
		ItemType:      b.ItemType,
		Item:          b.Item.Clone(),
		Index:         newIndex,
		Tag:           b.Tag,
		ElementString: b.ElementString,
		Actual:        b.Actual,
		Docs:          b.Docs,
		Comment:       b.Comment,
		Value:         b.Value,
		Elements:      nil,
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
