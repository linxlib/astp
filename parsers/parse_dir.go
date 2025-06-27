package parsers

import (
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/types"
	"os"
	"path/filepath"
)

// parseDir 解析一个目录
// 对于引用一个包的时候，直接解析其目录下的所有文件（不包含子目录）
func parseDir(dir string, proj *types.Project) map[string]*types.File {
	files := make(map[string]*types.File)
	fs, _ := os.ReadDir(dir)
	for _, f := range fs {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".go" {
			f1 := ParseFile(filepath.Join(dir, f.Name()), proj)
			files[f1.KeyHash] = f1
		}
	}
	// 分析完这个目录后, 进行其中类型标记为this的处理
	for _, file := range files {
		for _, s := range file.Struct {
			// 处理结构中的字段
			handleStructThisField(files, s)
			// 处理结构中的方法(参数和返回值)
			handleStructThisMethod(files, s)
		}
	}

	return files
}

func handleStructThisMethod(filesCopy map[string]*types.File, s *types.Struct) {
	for _, method := range s.Method {
		// 针对方法的参数
		for _, param := range method.Param {
			if param.Package != nil && param.Package.Type == constants.PackageSamePackage {
				for _, f := range filesCopy {
					for _, s2 := range f.Struct {
						if s2.Name == param.Type {
							if !s2.Top {
								handleStructThisField(filesCopy, s2)
							}
							param.Struct = s2.Clone()
							param.Package = s2.Package.Clone()
							break
						}
					}

				}
			}
			if param.Generic {
				for _, tp := range param.TypeParam {
					if tp.Package != nil && tp.Package.Type == constants.PackageSamePackage {
						for _, f := range filesCopy {
							for _, s2 := range f.Struct {
								if s2.Name == tp.Type {
									if !s2.Top {
										handleStructThisField(filesCopy, s2)
									}
									tp.Struct = s2.Clone()
									tp.Package = s2.Package.Clone()
									break
								}
							}

						}
					}
				}
			}

		}

		// 针对方法的返回值
		for _, result := range method.Result {
			if result.Package != nil && result.Package.Type == constants.PackageSamePackage {
				for _, f := range filesCopy {
					for _, s2 := range f.Struct {
						if s2.Name == result.Type {
							if !s2.Top {
								handleStructThisField(filesCopy, s2)
							}
							result.Struct = s2.Clone()
							result.Package = s2.Package.Clone()
						}
					}

				}
			}
			if result.Generic {
				for _, tp := range result.TypeParam {
					if tp.Package != nil && tp.Package.Type == constants.PackageSamePackage {
						for _, f := range filesCopy {
							for _, s2 := range f.Struct {
								if s2.Name == tp.Type {
									if !s2.Top {
										handleStructThisField(filesCopy, s2)
									}
									tp.Struct = s2.Clone()
									tp.Package = s2.Package.Clone()

								}
							}

						}
					}
				}
				// 还要处理result的
				if result.Struct != nil {
					for _, field := range result.Struct.Field {
						if field.Generic {
							field.Package = new(types.Package)
							field.Package.Type = constants.PackageSamePackage
						}
					}
				}

			}
		}

	}

}

func handleStructThisField(filesCopy map[string]*types.File, s *types.Struct) {
	for _, field := range s.Field {
		if field.Package != nil && field.Package.Type == constants.PackageSamePackage {
			for _, f := range filesCopy {
				for _, s2 := range f.Struct {
					if s2.Name == field.Type {
						if !s2.Top {
							handleStructThisField(filesCopy, s2)
						}
						field.Struct = s2.Clone()
						field.Package = s2.Package.Clone()
					}
				}
			}
		}
	}
	if s.Field != nil && len(s.Field) > 0 {
		s.Top = true
		for _, field := range s.Field {
			if field.Parent {
				s.Top = false
			}
		}
	}

}
