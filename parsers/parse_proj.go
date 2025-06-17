package parsers

import (
	"bufio"
	"fmt"
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"github.com/linxlib/astp/types"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ParseProj() *types.Project {
	modFile := "go.mod"
	modDir, _ := os.Getwd()
	modPkg := ""
	modPath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	sdkPath := filepath.Join(os.Getenv("GOROOT"), "src")
	modVersion := ""
	file, _ := os.Open(modFile)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(m, "module") {
			m = strings.TrimPrefix(m, "module")
			m = strings.TrimSpace(m)
			modPkg = m
		}
		if strings.HasPrefix(m, "go") {
			m = strings.TrimPrefix(m, "go")
			m = strings.TrimSpace(m)
			modVersion = m
			break
		}
	}
	proj := &types.Project{
		ModPkg:     modPkg,
		BaseDir:    modDir,
		ModName:    modPkg,
		ModVersion: modVersion,
		ModPath:    modPath,
		SdkPath:    sdkPath,
		Timestamp:  time.Now().Unix(),
		Generator:  "github.com/linxlib/astp",
		Version:    "v0.4",
	}
	ParseFile("./main.go", proj)

	AfterParseProj(proj)

	return proj
}

func AfterParseProj(proj *types.Project) {
	//先找到所有结构体中, 位于顶层的结构(即字段中没有嵌套的结构体/泛型), 加入到一个缓存map中, 后面使用
	var topStructs = make(map[string]*types.Struct)
	for _, file := range proj.File {
		for _, s := range file.Struct {
			if s.IsTop() {
				topStructs[internal.GetKeyHash(s.Package.Path, s.Name)] = s.Clone()
			}
		}
	}
	//找到所有结构体中(除了上面的结构之外), 各个内容含有PackageSamePackage的对象, 从上面缓存中直接拷贝一份覆盖之 => 处理 结构未解析 情况
	for _, file := range proj.File {
		for _, s := range file.Struct {
			if !s.IsTop() {
				if s.Package.Type == constants.PackageSamePackage {
					fmt.Printf("struct: %s is this package struct\n", s.Name)
				}
			}
			for _, tp := range s.TypeParam {
				if tp.Package == nil {
					if tp.Struct == nil {
						continue
					}
					if tp.Struct.Package == nil {
						panic(fmt.Sprintf("struct: %s is this package struct\n", s.Name))
					}
					if tp.Struct.Package.Type == constants.PackageSamePackage {
						fmt.Printf("struct: %s tp: %s is this package tp\n", s.Name, tp.Type)
					}
				} else {
					if tp.Package.Type == constants.PackageSamePackage {
						fmt.Printf("struct: %s tp: %s is this package tp\n", s.Name, tp.Type)
					}
				}

			}
			for _, field := range s.Field {
				if field.Package.Type == constants.PackageSamePackage {
					fmt.Printf("struct: %s field: [%s %s] is this package field\n", s.Name, field.Name, field.Type)
				}

				for _, tp := range field.TypeParam {
					if tp.Package == nil {
						if tp.Struct == nil {
							continue
						}
						if tp.Struct.Package == nil {
							panic(fmt.Sprintf("struct: %s is this package struct\n", s.Name))
						}
						if tp.Struct.Package.Type == constants.PackageSamePackage {
							fmt.Printf("struct: %s field: [%s %s] tp: %s is this package field\n", s.Name, field.Name, field.Type, tp.Type)
						}
					} else {
						if tp.Package.Type == constants.PackageSamePackage {
							fmt.Printf("struct: %s field: [%s %s] tp: %s is this package field\n", s.Name, field.Name, field.Type, tp.Type)
						}
					}

				}

			}
		}

	}

	//找到所有结构体中, 位于底层的结构(final, 没有其他结构嵌套它的). 拿到其已经经过上一步处理过的typeparam泛型类型内容
	//根据其parent字段, 向上查找, 拿到父级的字段/方法, 处理泛型的实际类型(final结构的typeparam), 再合并到当前结构 => 处理泛型继承情况

	//麻烦的点在于, 一开始解析时, 通过递归去解析, 但是解析完依然有很多类型是未解析(类型需要的其他类型未解析), 需要一种机制, 回头能更新之

	//针对proj中的文件, 拉出全部常量(elemType=enum), 再到当前文件中的Struct中查找对应类型, 改变该结构的数据

	for _, file := range proj.File {
		consts := make(map[string][]*types.Const)
		for _, c := range file.Const {
			if c.ElemType == constants.ElemEnum {
				consts[c.Type] = append(consts[c.Type], c)
			}
		}
		for k, vs := range consts {
			for _, s := range file.Struct {
				if s.Name == k {
					if s.Enum == nil {
						s.Enum = new(types.Enum)
						s.Enum.Enums = make([]*types.EnumItem, 0)
					}

					for idx, v := range vs {
						s.Enum.Type = s.Type
						s.Enum.TypeName = s.TypeName
						s.Enum.Private = s.Private
						s.Enum.Name = s.Name
						s.Enum.Enums = append(s.Enum.Enums, &types.EnumItem{
							Index:   idx,
							Name:    v.Name,
							Type:    v.Type,
							Value:   v.Value,
							Private: internal.IsPrivate(v.Name),
							Doc:     types.CopySlice(v.Doc),
							Comment: types.CopySlice(v.Comment),
						})
					}
				}
			}
		}
	}

}
