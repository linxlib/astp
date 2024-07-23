package internal

import (
	"go/ast"
	"strings"
)

func trimComment(line string) string {
	return strings.TrimLeft(strings.TrimLeft(line, "//"), " ")
}

func GetComment(cg *ast.CommentGroup) string {
	var result string
	if cg != nil && cg.List != nil {
		for _, comment := range cg.List {
			result += trimComment(comment.Text) + " "
		}
	}
	return result
}
func GetComments(cg *ast.CommentGroup) []string {
	var result []string
	if cg != nil && cg.List != nil {
		for _, comment := range cg.List {
			result = append(result, trimComment(comment.Text))
		}
	}
	return result
}

func GetDocs(cg *ast.CommentGroup) []string {
	var result []string
	if cg != nil && cg.List != nil {
		for _, comment := range cg.List {
			result = append(result, trimComment(comment.Text))
		}
	}
	return result
}

type CommandType int

const (
	TypeHttpMethod CommandType = 0 //http 请求方法
	TypeMiddleware CommandType = 1 //中间件类
	TypeOther      CommandType = 2 //其他
	TypeDoc        CommandType = 3 //注释内容
	TypeParam      CommandType = 4 //方法的参数和返回值专用
	TypeTagger     CommandType = 5 //这种类型仅用于标记一些元素
)

// DocCommand 注解命令
type DocCommand struct {
	Name  string
	Value string
	Type  CommandType
}

// ParseDoc 解析注解
func ParseDoc(doc []string, name string, i map[string]CommandType) []*DocCommand {
	docs := make([]*DocCommand, 0)
	if doc == nil {
		return docs
	}
	for _, s := range doc {
		if strings.HasPrefix(s, "@") {
			ps := strings.SplitN(s, " ", 2)
			value := ""
			if len(ps) == 2 {
				value = strings.TrimSpace(ps[1])
			}
			docName := strings.TrimLeft(ps[0], "@")
			docs = append(docs, &DocCommand{
				Name:  docName,
				Value: value,
				Type:  i[docName],
			})
		} else if strings.HasPrefix(s, name) {
			ps := strings.SplitN(s, " ", 2)
			value := ""
			if len(ps) == 2 {
				value = strings.TrimSpace(ps[1])
			}
			docs = append(docs, &DocCommand{
				Name:  name,
				Value: value,
				Type:  TypeOther,
			})
		} else {
			docs = append(docs, &DocCommand{
				Name:  name,
				Value: s,
				Type:  TypeOther,
			})
		}

	}
	return docs
}
