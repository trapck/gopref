package goscan

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/trapck/gopref/cfg"
	"github.com/trapck/gopref/fs"
	"github.com/trapck/gopref/model"
)

// Result is gofile scan result model
type Result struct {
	model.PkgFile
	PkgName      string
	ImportUsages []model.PkgImportUsage
}

// Scan analyses go file
func Scan(moduleName string, path string, i os.FileInfo, tempDir string) (*Result, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	res := Result{PkgFile: model.PkgFile{Path: strings.ReplaceAll(path, tempDir+"/", "")}}
	multiImport := false
	currentLine := 0
	currentFn := ""
	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) > 1 && parts[0] == cfg.PkgKey {
			res.PkgName = parts[1]
		} else if multiImport || (len(parts) > 1 && parts[0] == cfg.ImportKey) {
			processImportLine(moduleName, &multiImport, parts, &res.PkgFile)
		} else {
			usages := processRegularLine(&currentFn, currentLine, line, res.PkgName, &res.PkgFile)
			if len(usages) > 0 {
				res.ImportUsages = append(res.ImportUsages, usages...)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &res, nil
}

// ExtractModuleName extracts module name from file
func ExtractModuleName(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return "", fmt.Errorf("no lines in file")
	}
	parts := strings.Fields(scanner.Text())
	if len(parts) < 2 || parts[0] != cfg.ModuleKey {
		return "", fmt.Errorf("invalid module line %s", scanner.Text())
	}
	return parts[1], nil
}

func processImportLine(moduleName string, multiImport *bool, strParts []string, file *model.PkgFile) {
	if *multiImport {
		if len(strParts) > 0 {
			if strParts[0] == ")" {
				*multiImport = false
			} else {
				var i model.PkgImport
				if len(strParts) > 1 {
					i = newPkgImport(strParts[0], strParts[1])
				} else {
					i = newPkgImport(fs.LastPart(strParts[0]), strParts[0])
				}
				if strings.Contains(i.Path, moduleName) {
					file.Imports = append(file.Imports, i)
				}
			}
		}
	} else {
		if strParts[1] == "(" {
			*multiImport = true
		} else {
			var i model.PkgImport
			if len(strParts) > 2 {
				i = newPkgImport(strParts[1], strParts[2])
			} else {
				i = newPkgImport(fs.LastPart(strParts[1]), strParts[1])
			}
			if strings.Contains(i.Path, moduleName) {
				file.Imports = append(file.Imports, i)
			}
		}
	}
}

func processRegularLine(currentFn *string, lineNum int, line string, pkg string, file *model.PkgFile) []model.PkgImportUsage {
	usages := []model.PkgImportUsage{}
	parts := strings.Fields(line)
	if len(parts) > 0 {
		if parts[0] == "func" {
			*currentFn = strings.ReplaceAll(line, "{", "")
		}
		for _, v := range file.Imports {
			if strings.Contains(line, v.Alias+".") {
				usages = append(
					usages, model.PkgImportUsage{
						Line:       lineNum,
						Text:       line,
						Fn:         *currentFn,
						PkgImport:  v,
						InFilePath: file.Path,
						InPkgPath:  pkg,
					})
			}
		}
	}
	return usages
}

func newPkgImport(alias string, path string) model.PkgImport {
	return model.PkgImport{Alias: strings.ReplaceAll(alias, "\"", ""), Path: strings.ReplaceAll(path, "\"", "")}
}
