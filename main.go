package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type pkg struct {
	name string
	path string
}

type pkgFileInfo struct {
	pkg
	file    string
	imports []pkgImport
}

type pkgImport struct {
	pkg
	usages []pkgImportUsage
}

type pkgImportUsage struct {
	lineNum int
	line    string
	fn      string
}

type importedPkgInfo struct {
	pkg
	usedIn map[string][]importUsedIn
}

type importUsedIn struct {
	pkgFileInfo
	usages []pkgImportUsage
}

const (
	path   = "/Users/trapck/go/poc.bp.api"
	module = "poc.bp.api"
)

func main() {
	packages := map[string][]pkgFileInfo{}
	filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		fName := i.Name()
		if filepath.Ext(fName) == ".go" && !strings.Contains(fName, "_test") {
			pkgInfo, err := scanGo(p, i)
			if err != nil {
				log.Fatal(err)
			}
			appendPkg(packages, *pkgInfo)
		}
		return err
	})
	print(packages)
}

func appendPkg(packages map[string][]pkgFileInfo, info pkgFileInfo) {
	packageFiles, ok := packages[info.name]
	if !ok {
		packageFiles = []pkgFileInfo{}
	}
	packages[info.name] = append(packageFiles, info)
}

func scanGo(path string, i os.FileInfo) (*pkgFileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var pkgInfo pkgFileInfo
	multiImport := false
	currentLine := 0
	currentFn := ""
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) > 1 && parts[0] == "package" {
			pkgInfo = processPackageLine(path, parts, i)
		} else if multiImport || (len(parts) > 1 && parts[0] == "import") {
			processImportLine(&multiImport, parts, &pkgInfo)
		} else {
			processRegularLine(&currentFn, currentLine, line, &pkgInfo)
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &pkgInfo, nil
}

func processPackageLine(path string, strParts []string, i os.FileInfo) pkgFileInfo {
	return pkgFileInfo{pkg: pkg{name: strParts[1], path: path}, file: i.Name()}
}

func processImportLine(multiImport *bool, strParts []string, pkgInfo *pkgFileInfo) {
	if *multiImport {
		if len(strParts) > 0 {
			if strParts[0] == ")" {
				*multiImport = false
			} else {
				var i pkgImport
				if len(strParts) > 1 {
					i = newPkgImport(strParts[0], strParts[1])
				} else {
					importParts := strings.Split(strParts[0], "/")
					i = newPkgImport(importParts[len(importParts)-1], strParts[0])
				}
				pkgInfo.imports = append(pkgInfo.imports, i)
			}
		}
	} else {
		if strParts[1] == "(" {
			*multiImport = true
		} else {
			var i pkgImport
			if len(strParts) > 2 {
				i = newPkgImport(strParts[1], strParts[2])
			} else {
				importParts := strings.Split(strParts[1], "/")
				i = newPkgImport(importParts[len(importParts)-1], strParts[1])
			}
			pkgInfo.imports = append(pkgInfo.imports, i)
		}
	}
}

func processRegularLine(currentFn *string, lineNum int, line string, pkgInfo *pkgFileInfo) {
	parts := strings.Fields(line)
	if len(parts) > 0 {
		if parts[0] == "func" {
			*currentFn = strings.ReplaceAll(line, "{", "")
		}
		for i, v := range pkgInfo.imports {
			if strings.Contains(line, v.name+".") {
				pkgInfo.imports[i].usages = append(
					pkgInfo.imports[i].usages,
					pkgImportUsage{lineNum: lineNum, line: line, fn: *currentFn},
				)
			}
		}
	}
}

func newPkgImport(name string, path string) pkgImport {
	return pkgImport{pkg: pkg{name: strings.ReplaceAll(name, "\"", ""), path: strings.ReplaceAll(path, "\"", "")}}
}

func aggImports(pkgs map[string][]pkgFileInfo) map[string]importedPkgInfo {
	result := map[string]importedPkgInfo{}
	for _, p := range pkgs {
		for _, fileInPkg := range p {
			for _, i := range fileInPkg.imports {
				if _, ok := result[i.path]; !ok {
					result[i.path] = importedPkgInfo{pkg: i.pkg, usedIn: map[string][]importUsedIn{}}
				}
				result[i.path].usedIn[fileInPkg.name] = append(
					result[i.path].usedIn[fileInPkg.name],
					importUsedIn{pkgFileInfo: fileInPkg, usages: i.usages},
				)
			}
		}
	}
	return result
}

func print(packages map[string][]pkgFileInfo) {
	importInfo := aggImports(packages)
	//printPkgInfo(packages)
	//fmt.Println()
	//fmt.Println(strings.Repeat("*", 50))
	//fmt.Println()
	printImportInfo(importInfo)
}

func printPkgInfo(packages map[string][]pkgFileInfo) {
	for k, v := range packages {
		fmt.Println(k + strings.Repeat("=", 20))
		for _, v := range v {
			fmt.Println(fmt.Sprintf("%s%s %s", getIndent(1), v.file, v.path))
			for _, v := range v.imports {
				if strings.Contains(v.path, module) {
					fmt.Println(fmt.Sprintf("%s%s %s", getIndent(2), v.name, v.path))
					for _, v := range v.usages {
						fmt.Println(fmt.Sprintf("%s%d %s", getIndent(3), v.lineNum, v.fn))
						fmt.Println(fmt.Sprintf("%s%s", getIndent(4), v.line))
					}
				}
			}
		}
	}
}

func printImportInfo(packages map[string]importedPkgInfo) {
	keys := []string{}
	for k := range packages {
		if strings.Contains(k, module) {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		iParts := strings.Split(keys[i], "/")
		jParts := strings.Split(keys[j], "/")
		if len(iParts) != len(jParts) {
			return len(iParts) < len(jParts)
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		v := packages[k]
		fmt.Println(k + strings.Repeat("=", 20))
		for k, v := range v.usedIn {
			fmt.Println(fmt.Sprintf("%s%s", getIndent(1), k))
			for _, v := range v {
				fmt.Println(fmt.Sprintf("%s%s %s", getIndent(2), v.file, v.path))
				for _, v := range v.usages {
					fmt.Println(fmt.Sprintf("%s%d %s", getIndent(3), v.lineNum, v.fn))
					fmt.Println(fmt.Sprintf("%s%s", getIndent(4), v.line))
				}
			}
		}
	}
}

func getIndent(count int) string {
	return strings.Repeat("  ", count)
}
