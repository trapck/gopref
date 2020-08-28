package model

// Pkg is model of package
type Pkg struct {
	Module string    `bson:"module"`
	Name   string    `bson:"name"`
	Path   string    `bson:"path"`
	Files  []PkgFile `bson:"files"`
}

// PkgFile is model of package file
type PkgFile struct {
	Path    string      `bson:"path"`
	Imports []PkgImport `bson:"imports"`
}

// PkgImport is model of package's import
type PkgImport struct {
	Path  string `bson:"path"`
	Alias string `bson:"alias"`
}

// PkgImportUsage os model of package's import usage
type PkgImportUsage struct {
	Line       int    `bson:"line"`
	Text       string `bson:"text"`
	Fn         string `bson:"fn"`
	PkgImport  `bson:",inline"`
	InFilePath string `bson:"inFilePath"`
	InPkgPath  string `bson:"inPkgPath"`
}

// CombinedUsage is ... will be changed later
type CombinedUsage struct {
	Path   string
	InFile string
	Text   []string
}
