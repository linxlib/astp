package constants

type PackageType = string

const (
	PackageNormal       PackageType = "normal"
	PackageBuiltin      PackageType = "builtin"
	PackageSamePackage  PackageType = "this"
	PackageOtherPackage PackageType = "other"
	PackageThirdPackage PackageType = "third"
	PackageIgnore       PackageType = "ignore"
)
