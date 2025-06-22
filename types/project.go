package types

import (
	"compress/gzip"
	"encoding/json"
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"log/slog"
	"os"
	"slices"
	"strconv"
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
	jsonFileSize := strconv.FormatFloat(float64(len(jsonData))/1024.0, 'f', 2, 64)
	slog.Info("origin json file size", "size", jsonFileSize, "unit", "kb")
	slog.Info("project file count", "count", len(p.FileMap))
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
	fileInfo, _ := os.Stat(fileName)
	slog.Info("gzip compressed size", "size", strconv.FormatFloat(float64(fileInfo.Size())/1024.0, 'f', 2, 64), "unit", "kb")

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
			var delField *Field = nil
			for _, field := range s.Field {
				if field.Struct == nil || !field.Parent {
					continue
				}

				//slog.Info("[field's struct is this]", field.Struct.Name, field.Struct.Package.Path+"."+field.Struct.TypeName)
				//查找上级结构
				keyHash := internal.GetKeyHash(field.Struct.Package.Path, field.Struct.Type)
				newStruct := p.findStruct(keyHash)

				//newStruct.TypeParam = CopySlice(s.TypeParam)
				// 只要是隐式引用(没有Name),  均将上级结构的字段直接加入到本结构体 (json包就是这么处理的)
				if field.Name == constants.EmptyName {
					delField = field
					// 将上级结构的导出字段 加入到本结构体
					for _, f := range newStruct.Field {
						if !f.Private {
							s.Field = append(s.Field, f.Clone())
						}
					}
				} else {
					field.Struct = newStruct.Clone()
				}

				// 上级结构的注释中op=true的注释拷贝过来 其他注释无用
				for _, comment := range newStruct.Comment {
					if !comment.Op {
						continue
					}
					s.Comment = append(s.Comment, comment.Clone())
				}
				//将上级有标记为 @XXX 的方法加入到本结构体
				// 其他类型的方法没必要pull到本级
				// 这是为了实现Crud而做的处理
				for _, function := range newStruct.Method {
					if function.Private || !function.IsOp() {
						continue
					}
					cloned := function.Clone()
					for _, param := range cloned.Param {
						if !param.Generic {
							continue
						}
						var tpString []string
						handleParamTypeName(param)

						for _, typeParam := range newStruct.TypeParam {
							for idx, t := range param.TypeParam {
								if t.Type == typeParam.Type {
									clonedTp := typeParam.CloneTiny()
									handleTypeParamTypeName(clonedTp, t)
									pre := param.TypeParam[idx].Clone()
									clonedTp.Slice = pre.Slice
									clonedTp.Pointer = pre.Pointer
									param.TypeParam[idx] = clonedTp

									tpString = append(tpString, clonedTp.TypeName)
									//result.TypeParam[idx].Index = idx
								}
							}
						}
						// 将本级的泛型类型覆盖到参数这边
						for _, typeParam := range s.TypeParam {

							for idx, tp := range param.TypeParam {
								if tp.Index == typeParam.Index {
									pre := param.TypeParam[idx].Clone()
									cl := typeParam.Clone()
									cl.Slice = pre.Slice
									cl.Pointer = pre.Pointer
									param.TypeParam[idx] = cl

									param.Struct = typeParam.Struct.Clone()
									param.Package = typeParam.Struct.Package.Clone()
									param.Slice = pre.Slice
									param.Pointer = pre.Pointer
									continue
								}
							}
						}
						if len(tpString) > 0 {
							param.TypeName += "[" + strings.Join(tpString, ",") + "]"
						}
						if param.Struct != nil {
							for _, f := range param.Struct.Field {
								if f.Generic {
									for idx, typeParam := range param.Struct.TypeParam {
										if typeParam.Type == f.Type {
											for idx2, t := range param.TypeParam {
												if idx == idx2 {
													pre := f.Clone()
													f.Struct = t.Struct.Clone()
													f.Package = t.Struct.Package.Clone()
													f.Slice = pre.Slice
													f.Pointer = pre.Pointer
												}
											}

										}

									}
								} else {
									// 这里的field.Package 为 nil, 具体是哪里问题
									//fmt.Printf("字段为非泛型类型: %s.%s 并且为结构 需要专门处理\n", f.Package.Name, f.Name)
									if f.Package.Type == constants.PackageSamePackage {
										keyHash1 := internal.GetKeyHash(newStruct.Package.Path, f.Type)
										newStruct1 := p.findStruct(keyHash1)
										if newStruct1 != nil {
											f.Struct = newStruct1.Clone()
											f.Package = newStruct1.Package.Clone()

										}
									}

								}
							}
						}

						//param.TypeParam = make([]*TypeParam, 0)
						//for _, typeParam := range newStruct.TypeParam {
						//	if param.Type == typeParam.Type {
						//		param.TypeParam = append(param.TypeParam, typeParam.Clone())
						//	}
						//}
						//for _, typeParam := range s.TypeParam {
						//	for idx, tp := range param.TypeParam {
						//		if tp.Index == typeParam.Index {
						//			param.TypeParam[idx] = typeParam.Clone()
						//			param.Struct = typeParam.Struct.Clone()
						//			param.Slice = typeParam.Slice
						//			param.Pointer = typeParam.Pointer
						//			param.Type = param.Struct.Name
						//			continue
						//		}
						//	}
						//}
						//param.TypeName += "[" + strings.Join(tpString, ",") + "]"

					}

					for _, result := range cloned.Result {
						if !result.Generic {
							continue
						}
						var tpString []string
						handleParamTypeName(result)
						//if result.Slice {
						//	result.TypeName = "[]"
						//} else {
						//	result.TypeName = ""
						//}
						//if result.Pointer {
						//	result.TypeName += "*"
						//}
						//
						//result.TypeName += result.Package.Name + "." + result.Type
						//result.TypeParam = make([]*TypeParam, 0)

						//TODO: 处理方法时, 先将结构的

						// 先将上级的泛型类型覆盖到参数这边来(根据泛型参数索引)
						for _, typeParam := range newStruct.TypeParam {
							for idx, t := range result.TypeParam {
								if t.Type == typeParam.Type {
									clonedTp := typeParam.Clone()
									handleTypeParamTypeName(clonedTp, t)
									//clonedTp.TypeName = ""
									//if t.Slice {
									//	clonedTp.TypeName += "[]"
									//}
									//if t.Pointer {
									//	clonedTp.TypeName += "*"
									//}
									//clonedTp.TypeName += clonedTp.Struct.Package.Name + "." + clonedTp.Type
									pre := result.TypeParam[idx].Clone()
									clonedTp.Slice = pre.Slice
									clonedTp.Pointer = pre.Pointer
									result.TypeParam[idx] = clonedTp

									tpString = append(tpString, clonedTp.TypeName)
									//result.TypeParam[idx].Index = idx
								}
							}
						}
						// 将本级的泛型类型覆盖到参数这边
						for _, typeParam := range s.TypeParam {

							for idx, tp := range result.TypeParam {
								if tp.Index == typeParam.Index {
									//pre := result.TypeParam[idx].Clone()
									//
									//result.TypeParam[idx] = typeParam.Clone()
									//result.TypeParam[idx].Slice = pre.Slice
									//result.TypeParam[idx].Pointer = pre.Pointer

									pre := result.TypeParam[idx].Clone()
									cl := typeParam.Clone()
									cl.Slice = pre.Slice
									cl.Pointer = pre.Pointer
									result.TypeParam[idx] = cl

									//result.Struct = typeParam.Struct.Clone()
									//result.Package = typeParam.Struct.Package.Clone()
									//result.Slice = pre.Slice
									//result.Pointer = pre.Pointer
									continue
								}
							}
						}
						if len(tpString) > 0 {
							result.TypeName += "[" + strings.Join(tpString, ",") + "]"
						}

						// 处理参数结构中的字段
						if result.Struct != nil {
							for _, f := range result.Struct.Field {
								if f.Generic {
									for idx, typeParam := range result.Struct.TypeParam {
										if typeParam.Type == f.Type {

											for idx2, t := range result.TypeParam {
												if idx == idx2 {
													f.Struct = t.Struct.Clone()
													f.Package = t.Struct.Package.Clone()
													f.Slice = t.Slice
													f.Pointer = t.Pointer
												}
											}

										}

									}
								} else {
									// 这里的field.Package 为 nil, 具体是哪里问题
									//fmt.Printf("字段为非泛型类型: %s.%s 并且为结构 需要专门处理\n", f.Package.Name, f.Name)
									if f.Package.Type == constants.PackageSamePackage {
										keyHash1 := internal.GetKeyHash(newStruct.Package.Path, f.Type)
										newStruct1 := p.findStruct(keyHash1)
										if newStruct1 != nil {
											f.Struct = newStruct1.Clone()
											f.Package = newStruct1.Package.Clone()

										}
									}

								}
							}
						}

					}
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
			// 删除该字段
			if delField != nil {
				s.Field = slices.DeleteFunc(s.Field, func(i *Field) bool {
					return i.Name == delField.Name && i.Type == delField.Type && i.TypeName == delField.TypeName
				})
			}

			// 处理方法
			for _, method := range s.Method {
				if !method.IsOp() {
					continue
				}
				for _, param := range method.Param {
					if !param.Generic {
						continue
					}
					if param.Struct == nil {
						continue
					}
					for _, field := range param.Struct.Field {
						if !field.Generic {
							continue
						}
						for _, tp := range param.Struct.TypeParam {
							if tp.Type != field.Type {
								continue
							}
							field.Parent = true
							for _, tp1 := range param.TypeParam {
								if tp1.Index != tp.Index {
									continue
								}
								if field.TypeParam == nil {
									field.TypeParam = make([]*TypeParam, 0)
								}
								field.TypeParam = append(field.TypeParam, tp1.Clone())
								field.Slice = tp1.Slice
								field.Pointer = tp1.Pointer
								field.Struct = tp1.Struct.Clone()
								field.Package = tp1.Package.Clone()
								//TODO: field 内的struct->field 也要处理
								break
							}
							break
						}

					}

				}
				for _, param := range method.Result {
					if !param.Generic {
						continue
					}
					if param.Struct == nil {
						continue
					}
					for _, field := range param.Struct.Field {
						if !field.Generic {
							continue
						}
						for _, tp := range param.Struct.TypeParam {
							if tp.Type != field.Type {
								continue
							}
							field.Parent = true
							for _, tp1 := range param.TypeParam {
								if tp1.Index != tp.Index {
									continue
								}
								if field.TypeParam == nil {
									field.TypeParam = make([]*TypeParam, 0)
								}
								//field.TypeParam = append(field.TypeParam, tp1.Clone())
								field.Struct = tp1.Struct.Clone()
								field.Package = tp1.Package.Clone()
								break
							}
							break
						}

					}
				}

			}

		}
	}

}

func handleTypeParamTypeName(tp *TypeParam, t *TypeParam) {
	tp.TypeName = ""
	if t.Slice {
		tp.TypeName += "[]"
	}
	if t.Pointer {
		tp.TypeName += "*"
	}
	if t.Struct != nil && t.Struct.Package.Name != "" {
		tp.TypeName += t.Struct.Package.Name + "." + tp.Type
	} else {
		tp.TypeName += tp.Type
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

func (p *Project) FindStruct(keyHash string) *Struct {
	return p.findStruct(keyHash)
}

func handleParamTypeName(param *Param) {
	param.TypeName = ""

	if param.Slice {
		param.TypeName = "[]"
	}
	if param.Pointer {
		param.TypeName += "*"
	}
	if param.Package.Name == "" {
		param.TypeName += param.Type
	} else {
		param.TypeName += param.Package.Name + "." + param.Type
	}
}
