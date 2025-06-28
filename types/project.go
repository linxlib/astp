package types

import (
	"compress/gzip"
	"encoding/json"
	"github.com/linxlib/astp/constants"
	"github.com/linxlib/astp/internal"
	"log/slog"
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
	// Serialize project to JSON with indentation
	jsonData, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	//if os.Getenv("ASTP_DEBUG") == "true" {
	//
	//}
	jsonFileSize, _ := internal.Byte(len(jsonData)).ToString()
	slog.Info("origin json file", "size", jsonFileSize)
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

	// Write compressed data
	_, err = gz.Write(jsonData)
	defer func() {
		_ = gz.Close()
		fileInfo, _ := f.Stat()
		gzipFileSize, _ := internal.Byte(fileInfo.Size()).ToString()
		slog.Info("gzip compressed", "size", gzipFileSize)
	}()

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
	// 处理枚举合并(将常量合并到对应结构中, 仅合并同文件)
	p.handleEnum()
	for _, file := range p.FileMap {
		for _, s := range file.Struct {
			p.handleExistsMethods(s)
			p.handleAnonymousField(s)
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

func (p *Project) FindStruct(keyHash string) *Struct {
	return p.findStruct(keyHash)
}

func (p *Project) handleStructCurrentMethod(s *Struct) {
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
			// 泛型返回值
			if !param.Generic {
				continue
			}
			//有结构
			if param.Struct == nil {
				continue
			}
			for _, field := range param.Struct.Field {
				// 泛型字段
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
						field.Struct = tp1.Struct.Clone()
						field.Package = tp1.Package.Clone()

						handleGenericField(field.Struct, param.TypeParam, param.Struct, p)
						break
					}
					break
				}

			}
		}

	}
}

func handleNonGenericField(f *Field, parentStruct *Struct, p *Project) {
	if f.Package.Type == constants.PackageSamePackage {
		keyHash1 := internal.GetKeyHash(parentStruct.Package.Path, f.Type)
		newStruct1 := p.findStruct(keyHash1)
		if newStruct1 != nil {
			f.Struct = newStruct1.Clone()
			f.Package = newStruct1.Package.Clone()

		}
	}
}

func handleGenericField(s *Struct, tps []*TypeParam, parentStruct *Struct, p *Project) {
	s.VisitFields(func(f *Field) bool {
		return f.Generic
	}, func(f *Field) {
		for idx, typeParam := range s.TypeParam {
			if typeParam.Type == f.Type {
				for idx2, t := range tps {
					if idx == idx2 {
						//pre := f.Clone()
						newStruct := t.Struct.Clone()

						f.Struct = newStruct
						if f.Struct != nil {
							f.Package = t.Struct.Package.Clone()
						}

						f.Slice = t.Slice
						f.Pointer = t.Pointer
						//TODO: 如果这一级的f.Struct的字段列表中也有泛型类型, 则需要进一步处理. 为兼容多级, 需要递归处理

						handleGenericField2(f, parentStruct, p)
					}
				}

			}

		}
	})
	s.VisitFields(func(f *Field) bool {
		return !f.Generic
	}, func(f *Field) {
		handleNonGenericField(f, parentStruct, p)
	})
}
func handleGenericField2(s *Field, parentStruct *Struct, p *Project) {
	s.Struct.VisitFields(func(f *Field) bool {
		return f.Generic
	}, func(f *Field) {
		if len(s.TypeParam) == len(f.TypeParam) {
			for idx, typeParam := range s.TypeParam {
				//pre := f.TypeParam[idx].Clone()
				newTp := typeParam.Clone()
				f.TypeParam[idx] = newTp
				newStruct := newTp.Struct.Clone()
				if newStruct != nil {
					f.Package = newStruct.Package.Clone()
					f.Struct = newStruct
				}
				f.Slice = s.Slice
				f.Pointer = s.Pointer
				handleGenericField2(f, parentStruct, p)
			}
		}

	})
	s.Struct.VisitFields(func(f *Field) bool {
		return !f.Generic
	}, func(f *Field) {
		handleNonGenericField(f, parentStruct, p)
	})
}

// handleParentParam 处理泛型参数/返回值
// currentStruct: 当前结构
// param: 参数
// tps: 类型参数(匿名字段对应的结构)
func (p *Project) handleParentParam(currentStruct *Struct, param *Param, parentStructTypeParams []*TypeParam) {
	if !param.Generic {
		if param.Struct != nil {
			for _, field := range param.Struct.Field {
				p.handleNonGenericField(field, param.Struct)
			}
		}

		return
	}
	if param.Struct == nil {
		// 1. 参数本身的TypeParam eg. param *E / (*E, error) / []*E / ([]*E, error)
		for _, typeParam := range currentStruct.TypeParam {
			if param.Type == typeParam.OType {
				param.Struct = typeParam.Struct.Clone()
			}
		}
		//for _, tp := range parentStructTypeParams {
		//	for _, typeParam := range param.TypeParam {
		//		if typeParam.Key == tp.Key { // 匹配到了
		//			// 设置索引, 表示参数的这个泛型对应结构中的第几个(后续用于覆盖真实的泛型类型)
		//			typeParam.Index = tp.Index
		//
		//			for _, t := range currentStruct.TypeParam {
		//				if t.Index == typeParam.Index { // 真实的泛型类型 需要以索引(第几个泛型参数)来匹配
		//					typeParam.Struct = t.Struct.Clone()
		//					typeParam.Package = t.Package.Clone()
		//					param.Struct = t.Struct.Clone()
		//				}
		//			}
		//
		//		}
		//	}
		//}
		return
	}

	// 循环处理
	// 2. 参数结构的TypeParam和其字段 eg. *resp.Resp[*E] / *resp.Resp[[]*E]
	// 3. 多层嵌套型 eg. *resp.Resp[PageResult[[]*E]]

	// 处理参数结构的各个字段, 如果字段仍然为泛型, 则递归处理
	p.handleParentMethodParamGeneric(param.Struct, param.TypeParam, currentStruct.TypeParam)

	p.handleParentMethodParamRealGeneric(param.Struct, param.TypeParam, currentStruct.TypeParam)
	//slog.Info("handleParentMethodParamRealGeneric", "typeName", param.TypeName)
	tn := strings.Builder{}
	if param.Slice {
		tn.WriteString("[]")
	}
	if param.Pointer {
		tn.WriteString("*")
	}
	tn.WriteString(param.Struct.TypeName) //*param.Resp
	var tpString []string
	tn1 := strings.Builder{}
	if param.Struct != nil {
		for _, field := range param.Struct.Field {
			if !field.Generic {
				continue
			}
			tn1.WriteString(field.TypeName)
			var tpString1 []string
			if field.Struct != nil {
				for _, f := range field.Struct.Field {
					if !f.Generic {
						continue
					}
					tpString1 = append(tpString1, f.TypeName)

				}
			}
			if len(tpString1) > 0 {
				tn1.WriteString("[" + strings.Join(tpString1, ",") + "]")
			}
			tpString = append(tpString, tn1.String())
		}
	}
	if len(tpString) > 0 {
		tn.WriteString("[" + strings.Join(tpString, ",") + "]")
	}
	param.TypeName = tn.String()
}

func (p *Project) handleParam(currentStruct *Struct, param *Param) {
	if !param.Generic {
		if param.Struct != nil {
			for _, field := range param.Struct.Field {
				p.handleNonGenericField(field, param.Struct)
			}
		}

		return
	}
	// 循环处理
	// 1. 参数本身的TypeParam eg. param *E / (*E, error) / []*E / ([]*E, error)
	// 2. 参数结构的TypeParam和其字段 eg. *resp.Resp[*E] / *resp.Resp[[]*E]
	// 3. 多层嵌套型 eg. *resp.Resp[PageResult[[]*E]]

	// 递归查找, 更新泛型的Index(代表对应 currentStruct 泛型中的第几个)
	p.handleMethodParamGeneric(param.Struct, param.TypeParam, currentStruct.TypeParam)

	p.handleMethodParamRealGeneric(param.Struct, param.TypeParam)
}

func (p *Project) handleParentMethodParamRealGeneric(s *Struct, paramTypeParams []*TypeParam, tps []*TypeParam) {
	//slog.Info("handleParentMethodParamRealGeneric")
	if s == nil {
		return
	}
	for idx0, tp0 := range s.TypeParam {
		for _, field := range s.Field {
			if !field.Generic {
				continue
			}
			//slog.Info("handleParentMethodParamRealGeneric", "field", field.Name)
			for _, param := range paramTypeParams {
				for _, typeParam := range field.TypeParam {
					if typeParam.Index == param.Index {
						if param.ElemType == constants.ElemStruct {
							field.Struct = param.Struct.Clone()
							field.Pointer = param.Pointer
							field.Slice = param.Slice
							if field.Pointer {
								tmp := field.Struct.TypeName
								tmp = "*" + tmp
								if field.Slice {
									tmp = "[]" + tmp
								}
								field.TypeName = tmp
							} else {
								tmp := field.Struct.TypeName
								if field.Slice {
									tmp = "[]" + tmp
								}
								field.TypeName = tmp
							}
							p.handleParentMethodParamRealGeneric(field.Struct, field.Struct.TypeParam, tps)
						} else {
							for idx2, typeParam1 := range field.TypeParam {
								if typeParam1.Type != field.Type {
									continue
								}
								for _, tp := range tps {
									if tp.Index == typeParam1.Index {
										cloned := tp.Clone()
										cloned.Pointer = tp0.Pointer
										cloned.Slice = tp0.Slice
										field.TypeParam[idx2] = cloned
										field.Pointer = tp0.Pointer
										field.Slice = tp0.Slice
										field.Struct = tp.Struct.Clone()
										if field.Pointer {
											tmp := field.Struct.TypeName
											tmp = "*" + tmp
											if field.Slice {
												tmp = "[]" + tmp
											}
											field.TypeName = tmp
										} else {
											tmp := field.Struct.TypeName
											if field.Slice {
												tmp = "[]" + tmp
											}
											field.TypeName = tmp
										}
										s.TypeParam[idx0] = tp.Clone()
										s.TypeParam[idx0].Pointer = field.Pointer
										s.TypeParam[idx0].Slice = field.Slice
										//TODO: 最外层的param的TypeName需要更新
										break
									}
								}
							}
						}
					}
				}
			}

		}
	}
}

func (p *Project) handleMethodParamRealGeneric(s *Struct, tps []*TypeParam) {
	//slog.Info("handleMethodParamRealGeneric")
	if s == nil {
		return
	}
	for _, param := range s.TypeParam {
		if param.Index == -1 {
			for _, field := range s.Field {
				//slog.Info("handleMethodParamRealGeneric", "field", field.Name)
				has := false
				for idx2, typeParam := range field.TypeParam {
					if typeParam.Index != -1 {
						has = true
						for _, tp := range tps {
							if tp.Index == typeParam.Index {
								field.TypeParam[idx2] = tp.Clone()
								field.Struct = tp.Struct.Clone()
								if field.Pointer {
									tmp := field.Struct.TypeName
									tmp = "*" + tmp
									if field.Slice {
										tmp = "[]" + tmp
									}
									field.TypeName = tmp
								} else {
									tmp := field.Struct.TypeName
									if field.Slice {
										tmp = "[]" + tmp
									}
									field.TypeName = tmp
								}
								//TODO: 最外层的param的TypeName需要更新
								break
							}
						}
					}
				}
				if !has {
					p.handleMethodParamRealGeneric(field.Struct, tps)
				}

			}
		}
	}
}

// handleGeneric 处理泛型(递归)
func (p *Project) handleParentMethodParamGeneric(paramStruct *Struct, paramTypeParams []*TypeParam, parentStructTypeParams []*TypeParam) {
	if paramStruct == nil {
		return
	}
	//slog.Info("handleGeneric", "struct", paramStruct.Name)
	has := false
	for _, pstp := range parentStructTypeParams {
		for idx, ptp := range paramTypeParams {
			if ptp.Type == pstp.OType {
				has = true
				for _, typeParam := range paramStruct.TypeParam {
					if typeParam.Index == ptp.Index {
						clonedTypeParam := ptp.Clone()
						for _, field := range paramStruct.Field {
							if !field.Generic {
								continue
							}

							field.Struct = pstp.Struct.Clone()

							field.Slice = clonedTypeParam.Slice
							field.Pointer = clonedTypeParam.Pointer
							if clonedTypeParam.Slice {
								field.Slice = clonedTypeParam.Slice
							}
							if clonedTypeParam.Pointer {
								field.Pointer = clonedTypeParam.Pointer
							}
							field.TypeParam = make([]*TypeParam, 0)
							field.TypeParam = append(field.TypeParam, clonedTypeParam)
							field.TypeName = clonedTypeParam.TypeName
							field.Type = clonedTypeParam.Type
							if field.Struct != nil {
								p.handleParentMethodParamGeneric(field.Struct, field.TypeParam, parentStructTypeParams)
							}
							clonedTypeParam.Index = pstp.Index // 更新参数的泛型参数的Index
							clonedTypeParam.Key = pstp.Key
							paramStruct.TypeParam[idx] = clonedTypeParam
						}
					}
				}

			}
		}
	}
	if !has {
		for _, field := range paramStruct.Field {
			if field.Generic {
				p.handleParentMethodParamGeneric(field.Struct, paramStruct.TypeParam, parentStructTypeParams)
			}
		}
	}

}
func (p *Project) handleMethodParamGeneric(s *Struct, tps []*TypeParam, thatTps []*TypeParam) {
	if s == nil {
		return
	}
	//slog.Info("handleGeneric", "struct", s.Name)
	for idx, _ := range s.TypeParam {
		for idx1, t := range tps {
			if idx == idx1 {
				clonedTypeParam := t.Clone()
				for _, field := range s.Field {
					if field.Generic {
						field.Struct = t.Struct.Clone()
						if field.Struct != nil {
							clonedTypeParam.Index = -1
							clonedTypeParam.TypeInterface = "nonce"
						}
						field.TypeParam = make([]*TypeParam, 0)
						field.TypeParam = append(field.TypeParam, clonedTypeParam)
						field.Slice = clonedTypeParam.Slice
						field.Pointer = clonedTypeParam.Pointer
						if field.Slice {
							tmp := field.TypeName
							if field.Pointer {
								tmp = "*" + tmp
							}
							field.TypeName = "[]" + tmp
						}

						if field.Struct != nil {
							p.handleMethodParamGeneric(field.Struct, field.TypeParam, thatTps)
						}
						cloneT := t.Clone()
						cloneT.Index = -1
						s.TypeParam[idx] = cloneT
					}

				}
			}
		}

	}
	//slog.Info("handleMethodParamGeneric", "typeParam", s.TypeParam)

}

func (p *Project) handleParentMethod(currentStruct *Struct, parentMethod *Function, parentStructTypeParams []*TypeParam) {
	// 处理receiver
	// 替换拷贝过来的方法的接收器为当前类型
	parentMethod.Receiver = nil
	if currentStruct.Method != nil && len(currentStruct.Method) > 0 {
		parentMethod.Receiver = currentStruct.Method[0].Receiver.Clone()
	}

	//处理 param
	for _, param := range parentMethod.Param {
		p.handleParentParam(currentStruct, param, parentStructTypeParams)
	}
	//处理result
	for _, param := range parentMethod.Result {
		p.handleParentParam(currentStruct, param, parentStructTypeParams)
	}
}

func (p *Project) handleAnonymousField(currentStruct *Struct) {
	var delField *Field = nil
	for _, field := range currentStruct.Field {
		if field.Struct == nil || !field.Parent {
			continue
		}
		//先查找字段对应的结构
		keyHash := internal.GetKeyHash(field.Struct.Package.Path, field.Struct.Type)
		fieldStruct := p.findStruct(keyHash).CloneFull()

		// 只要是隐式引用(没有Name),  均将上级结构的字段直接加入到本结构体 (json包就是这么处理的)
		if field.Name == constants.EmptyName {
			delField = field
			currentStruct.TypeParam = CopySlice(field.TypeParam)
			// 由于字段是隐式引用 (没有Name), 将该结构的字段添加到当前结构的字段列表中
			for _, f := range fieldStruct.Field {
				if !f.Private {
					currentStruct.Field = append(currentStruct.Field, f.Clone())
				}
			}
		} else {
			// 对于已经经过this处理的结构, 这里更新一下.
			// 字段对应的结构中可能包含this, 而this的处理在 parse_dir.go:22 进行过
			// 这里针对的一般是 other 类型
			field.Struct = fieldStruct
		}

		// 上级结构的注释中op=true的注释拷贝过来 其他注释无用
		// 即 类似 @XXX 这样的注解
		for _, comment := range fieldStruct.Doc {
			if !comment.Op {
				continue
			}
			currentStruct.Comment = append(currentStruct.Comment, comment.Clone())
		}

		// 提取上级结构的泛型类型定义(这里一般还是T,E这样的'Type')
		// 用于后面receiver param result的处理
		tps := CopySlice(fieldStruct.TypeParam)
		for _, tp := range tps {
			for _, param := range currentStruct.TypeParam {
				if tp.Index == param.Index {
					param.OType = tp.Type
				}
			}
		}

		for _, function := range fieldStruct.Method {
			// 跳过私有方法和非操作方法
			if function.Private || !function.IsOp() {
				continue
			}
			cloned := function.Clone()
			//slog.Info("handleParentMethod", "method", cloned.Name)
			p.handleParentMethod(currentStruct, cloned, tps)
			currentStruct.Method = append(currentStruct.Method, cloned)
		}
		// 简单排序一下
		slices.SortFunc(currentStruct.Method, func(a, b *Function) int {
			return a.Index - b.Index
		})

	}
	// 对于隐式引用(没有Name)的字段, 删除它, 对后面没什么用了
	if delField != nil {
		// 也可以不删除, 而是将字段相关设置为非公开字段等
		currentStruct.Field = slices.DeleteFunc(currentStruct.Field, func(i *Field) bool {
			return i.Name == delField.Name && i.Type == delField.Type && i.TypeName == delField.TypeName
		})
	}
}

func (p *Project) handleExistsMethods(currentStruct *Struct) {
	isOp := false
	for _, comment := range currentStruct.Doc {
		if !comment.Op {
			continue
		}
		isOp = true
		break
	}
	if !isOp {
		return
	}
	for _, method := range currentStruct.Method {
		if method.Private || !method.IsOp() {
			continue
		}
		//处理 param
		for _, param := range method.Param {
			p.handleParam(currentStruct, param)
		}
		//处理result
		for _, param := range method.Result {
			p.handleParam(currentStruct, param)
		}
	}

}

func (p *Project) handleNonGenericField(f *Field, parentStruct *Struct) {
	if f.Package.Type == constants.PackageSamePackage {
		keyHash1 := internal.GetKeyHash(parentStruct.Package.Path, f.Type)
		newStruct1 := p.findStruct(keyHash1)
		if newStruct1 != nil {
			f.Struct = newStruct1.Clone()
			f.Package = newStruct1.Package.Clone()
		}
	}
}
