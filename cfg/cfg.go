package cfg

import "time"

// App settings constants
const (
	GithubURLTemplate = "https://github.com/%s/%s/archive/master.zip"
	Tmp               = "/Users/trapck/go/gopref/temp"
	ExcludeTest       = true
	MongoHost         = "mongodb://localhost:27017"
	MongoDBName       = "gopkg"
	MongoDefTimeout   = 5 * time.Second
	Port              = 3000
	ModuleFile        = "go.mod"
	GoExt             = ".go"
	TestFileMask      = "_test"
	PkgKey            = "package"
	ImportKey         = "import"
	ModuleKey         = "module"
)
