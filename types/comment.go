package types

import (
	"fmt"
	"github.com/linxlib/astp/constants"
	"regexp"
	"strings"
)

var _ IElem[*Comment] = (*Comment)(nil)

type Comment struct {
	Index      int                `json:"index"`
	Content    string             `json:"content"`           //原始的一行注释内容
	IsSelf     bool               `json:"is_self,omitempty"` // 注释是否是以自己名称开头
	Op         bool               `json:"op,omitempty"`
	AttrType   constants.AttrType `json:"attr_type"`
	CustomAttr string             `json:"custom_attr,omitempty"`
	AttrValue  string             `json:"attr_value,omitempty"`
}

func (c *Comment) String() string {
	return fmt.Sprintf("Comment{Index: %d, Content: %s}", c.Index, c.Content)
}

func (c *Comment) Clone() *Comment {
	if c == nil {
		return nil
	}
	return &Comment{
		Index:      c.Index,
		Content:    c.Content,
		Op:         c.Op,
		IsSelf:     c.IsSelf,
		AttrType:   c.AttrType,
		CustomAttr: c.CustomAttr,
		AttrValue:  c.AttrValue,
	}
}

func (c *Comment) GetWithoutSelf(name string) string {
	return strings.TrimPrefix(strings.TrimSpace(c.Content), name)
}

func OfComment(index int, content string, selfName string) *Comment {
	var attrType = constants.AT_NONE
	attrCustom := ""
	attrValue := ""
	var op = false
	isSelf := false
	if strings.HasPrefix(content, "@") {
		re := regexp.MustCompile(`@(\S+)`)
		matches := re.FindAllStringSubmatch(content, -1)
		tmp0 := "@" + matches[0][1]
		tmp := strings.ToUpper(matches[0][1])
		op = true
		if v, ok := constants.AttrTypes[tmp]; ok {
			attrType = v
		} else {
			attrType = constants.AT_CUSTOM
			attrCustom = tmp
		}
		a := strings.TrimSpace(content)
		a = strings.TrimPrefix(a, tmp0)
		a = strings.TrimSpace(a)
		attrValue = a
	} else {
		if strings.HasPrefix(strings.TrimSpace(content), selfName) {
			isSelf = true
			a := strings.TrimSpace(content)
			a = strings.TrimPrefix(a, selfName)
			a = strings.TrimSpace(a)
			attrValue = a
		}
	}

	return &Comment{
		Index:      index,
		Content:    content,
		Op:         op,
		IsSelf:     isSelf,
		AttrType:   attrType,
		CustomAttr: attrCustom,
		AttrValue:  attrValue,
	}
}
func (c *Comment) IsHttpMethod() bool {

	return c.Op && (c.AttrType == constants.AT_ANY ||
		c.AttrType == constants.AT_POST ||
		c.AttrType == constants.AT_PUT ||
		c.AttrType == constants.AT_DELETE ||
		c.AttrType == constants.AT_GET || c.AttrType == constants.AT_OPTIONS || c.AttrType == constants.AT_PATCH || c.AttrType == constants.AT_HEAD)
}
