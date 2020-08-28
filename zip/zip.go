package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/trapck/gopref/goscan"

	"github.com/trapck/gopref/cfg"
)

// Result is an extract go repo archive result
type Result struct {
	Root   string
	Module string
}

// Unzip extracts files from src to dest folder
func Unzip(src string, dest string) (Result, error) {
	result := Result{Root: dest}
	r, err := zip.OpenReader(src)
	if err != nil {
		return result, err
	}
	defer r.Close()

	for i, f := range r.File {
		fi := f.FileInfo()
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return result, fmt.Errorf("%s: illegal file path", fpath)
		}
		if fi.IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			if i == 0 {
				result.Root = fpath
			}
			continue
		}
		if strings.Contains(f.Name, cfg.ModuleFile) {
			rc, err := f.Open()
			if err != nil {
				return result, err
			}
			defer rc.Close()
			mn, err := goscan.ExtractModuleName(rc)
			if err != nil {
				return result, err
			}
			result.Module = mn
			continue
		}
		if filepath.Ext(f.Name) != cfg.GoExt || (cfg.ExcludeTest && strings.Contains(f.Name, cfg.TestFileMask)) {
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return result, err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		defer outFile.Close()
		if err != nil {
			return result, err
		}
		rc, err := f.Open()
		if err != nil {
			return result, err
		}
		defer rc.Close()
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return result, err
		}
	}
	if result.Module == "" {
		return result, fmt.Errorf("unable to detect module name. go.mod file is required")
	}
	return result, nil
}
