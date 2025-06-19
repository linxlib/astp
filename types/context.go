package types

type Context struct {
	Proj    *Project
	Package *Package
	Imports []*Import
}
