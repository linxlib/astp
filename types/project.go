package types

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"os"
	"slices"
	"strings"
)

type Project struct {
	ModPkg     string           `json:"mod_pkg,omitempty"`
	BaseDir    string           `json:"base_dir,omitempty"`
	ModName    string           `json:"mod_name,omitempty"`
	ModVersion string           `json:"mod_version,omitempty"`
	ModPath    string           `json:"mod_path,omitempty"`
	SdkPath    string           `json:"sdk_path,omitempty"`
	Timestamp  int64            `json:"timestamp,omitempty"`
	Generator  string           `json:"generator,omitempty"`
	Version    string           `json:"version,omitempty"`
	FileMap    map[string]*File `json:"file,omitempty"`
	//File       []*File          `json:"file,omitempty"`
}

func (p *Project) AddFile(f *File) {
	if p.FileMap == nil {
		p.FileMap = make(map[string]*File)
	}
	p.FileMap[f.KeyHash] = f
}

func (p *Project) Merge(files map[string]*File) {
	for key, file := range files {
		if _, ok := p.FileMap[key]; !ok {
			p.FileMap[key] = file.Clone()
		}
	}
}

func (p *Project) Write(fileName string) error {
	//// 暂定: 不存储 FileMap
	//// 后续考虑fw框架是否需要在文件中进行快速查询
	//// 类似 Struct 也类似, 如果需要快速查询, 则全部改造为map 而不是slice
	//p.File = make([]*File, 0)
	//for _, file := range p.FileMap {
	//	p.File = append(p.File, file)
	//}
	// TODO: 也可以考虑使用gob进行序列化
	// Serialize project to JSON with indentation
	jsonData, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	//if os.Getenv("ASTP_DEBUG") == "true" {
	//
	//}
	jsonPath := strings.ReplaceAll(fileName, ".gz", ".json")
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		return err
	}

	// 对json数据进行gzip压缩, 减小文件体积
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create gzip writer
	gz := gzip.NewWriter(f)
	defer gz.Close()

	// Write compressed data
	_, err = gz.Write(jsonData)
	return err
}
func (p *Project) Read(path string) error {
	// Open the gzipped file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a gzip reader
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	// Decode the JSON data into the Project struct
	if err := json.NewDecoder(gz).Decode(p); err != nil {
		return err
	}

	return nil
}

// handleEnum 处理枚举
// 提取文件常量(elemType=enum), 并在当前文件中查找对应结构, 扩充该结构的enum字段
func (p *Project) handleEnum() {
	// 将枚举合并当当前文件的结构中去
	// 这里仅处理单个文件
	for _, file := range p.FileMap {
		enums := make(map[string][]*Const)
		for _, c := range file.Const {
			if c.ElemType == constants.ElemEnum {
				enums[c.Type] = append(enums[c.Type], c)
			}
		}
		for k, vs := range enums {
			for _, s := range file.Struct {
				if s.Name == k {
					if s.Enum == nil {
						s.Enum = new(Enum)
						s.Enum.Enums = make([]*EnumItem, 0)
					}

					for idx, v := range vs {
						s.Enum.Type = s.Type
						s.Enum.TypeName = s.TypeName
						s.Enum.Private = s.Private
						s.Enum.Name = s.Name
						s.Enum.ElemType = constants.ElemEnum
						s.Enum.Iota = v.Iota
						s.Enum.Enums = append(s.Enum.Enums, &EnumItem{
							Index: idx,
							Name:  v.Name,
							Type:  v.Type,
							Value: v.Value,

							Private: internal.IsPrivate(v.Name),
							Doc:     CopySlice(v.Doc),
							Comment: CopySlice(v.Comment),
						})
					}
				}
			}
		}
	}
}

func (p *Project) AfterParseProj() {
	p.handleEnum()
	// 由于解析顺序问题, 前期解析时, 对于类型就在当前文件的情况, 标记为了 Package.Type = "this"
	// 类似这样的结构, 需要在全部文件都解析后, 再进行处理
	for _, file := range p.FileMap {
		for _, s := range file.Struct {
			// 处理结构的字段
			for _, field := range s.Field {
				if field.Package.Type == constants.PackageSamePackage {
					keyHash := internal.GetKeyHash(file.Package.Path, field.Type)
					field.Struct = file.FindStruct(keyHash).Clone()
					field.Package = field.Struct.Package.Clone()

					if field.Struct != nil && field.Parent {
						keyHash := internal.GetKeyHash(field.Struct.Package.Path, field.Struct.Type)
						newStruct := p.findStruct(keyHash)

						// 将上级结构的导出字段 加入到本结构体
						for _, f := range newStruct.Field {
							if !f.Private {
								s.Field = append(s.Field, f.Clone())
							}
						}
					}
					continue
				}

			}

			// 处理结构的方法
			for _, function := range s.Method {
				// 对于结构方法的参数进行处理
				for _, param := range function.Param {
					if param.Package.Type == constants.PackageSamePackage {
						keyHash := internal.GetKeyHash(s.Package.Path, param.Type)
						param.Struct = file.FindStruct(keyHash).Clone()
						param.Package = param.Struct.Package.Clone()
						continue
					}
					if param.Generic {

						param.TypeParam = make([]*TypeParam, 0)
						for _, typeParam := range s.TypeParam {
							if param.Type == typeParam.Type {
								param.TypeParam = append(param.TypeParam, typeParam.Clone())
							}

						}

					}
				}
				// 对于结构方法的返回值进行处理
				for _, param := range function.Result {
					if param.Package.Type == constants.PackageSamePackage {
						keyHash := internal.GetKeyHash(s.Package.Path, param.Type)
						param.Struct = file.FindStruct(keyHash).Clone()
						if param.Struct == nil {
							continue
						}
						param.Package = param.Struct.Package.Clone()
						continue
					}
					//参数/返回值 是泛型参数
					if param.Generic {
						param.TypeParam = make([]*TypeParam, 0)
						for _, typeParam := range s.TypeParam {
							if param.Type == typeParam.Type {
								param.TypeParam = append(param.TypeParam, typeParam.Clone())
							}

						}
					}
				}

			}
		}
	}

	for _, file := range p.FileMap {
		for _, s := range file.Struct {
			if !s.IsTop() {

				for _, field := range s.Field {
					if field.Struct != nil && field.Parent {
						// 已经解析了其中 this 包的部分
						// 当前 field.Struct 还是老的结构信息
						keyHash := internal.GetKeyHash(field.Struct.Package.Path, field.Struct.Type)
						newStruct := p.findStruct(keyHash)

						// 上级结构的注释中op=true的注释拷贝过来 其他注释无用
						for _, comment := range newStruct.Comment {
							if comment.Op {
								s.Comment = append(s.Comment, comment.Clone())
							}
						}

						// 将上级结构的导出字段 加入到本结构体
						//for _, f := range newStruct.Field {
						//	if !f.Private {
						//		s.Field = append(s.Field, f.Clone())
						//	}
						//}

						//将上级有标记为 @XXX 的方法加入到本结构体
						// 其他类型的方法没必要pull到本级
						// 这是为了实现Crud而做的处理
						for _, function := range newStruct.Method {
							if !function.Private && function.IsOp() {
								cloned := function.Clone()
								for _, param := range cloned.Param {
									if param.Generic {
										for _, typeParam := range param.TypeParam {
											for _, typeParam1 := range s.TypeParam {
												if typeParam1.Index == typeParam.Index {
													keyHash1 := internal.GetKeyHash(typeParam1.Struct.Package.Path, typeParam1.Struct.Type)
													newStruct1 := p.findStruct(keyHash1)
													param.Struct = newStruct1.Clone()
												}
											}
										}
									}
								}
								for _, param := range cloned.Result {
									if param.Generic {
										for _, typeParam := range param.TypeParam {
											for _, typeParam1 := range s.TypeParam {
												if typeParam1.Index == typeParam.Index {
													keyHash1 := internal.GetKeyHash(typeParam1.Struct.Package.Path, typeParam1.Struct.Type)
													newStruct1 := p.findStruct(keyHash1)
													param.Struct = newStruct1.Clone()
												}
											}
										}

									}
								}
								//cloned.Receiver = nil
								// 对于方法的注释, Receiver需要替换
								//
								cloned.Receiver = nil
								if s.Method != nil && len(s.Method) > 0 {
									cloned.Receiver = s.Method[0].Receiver.Clone()
								}

								s.Method = append(s.Method, cloned)
								slices.SortFunc(s.Method, func(a, b *Function) int {
									return a.Index - b.Index
								})
							}
						}

					}

				}
				if s.Package.Type == constants.PackageSamePackage {
					fmt.Printf("struct: %s is this package struct\n", s.Name)

				}
			}

		}

	}
	for _, file := range p.FileMap {
		for _, s := range file.Struct {
			for _, function := range s.Method {
				for _, param := range function.Param {
					if param.Package.Type == constants.PackageSamePackage {
						//fmt.Printf("struct: %s method: %s param:( %s %s) is this package struct\n", s.Name, function.Name, param.Name, param.Type)
						keyHash := internal.GetKeyHash(s.Package.Path, param.Type)
						param.Struct = file.FindStruct(keyHash).Clone()
						param.Package = param.Struct.Package.Clone()
					}
				}
			}
		}
	}

}

func (p *Project) findStruct(keyHash string) *Struct {
	for _, f := range p.FileMap {
		if s := f.FindStruct(keyHash); s != nil {
			return s
		}
	}
	return nil
}
