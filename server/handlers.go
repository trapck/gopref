package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber"
	"github.com/trapck/gopref/cfg"
	"github.com/trapck/gopref/fs"
	"github.com/trapck/gopref/goscan"
	"github.com/trapck/gopref/model"
	"github.com/trapck/gopref/remote"
	"github.com/trapck/gopref/zip"
)

// HandlePackages handles imports info get request
func (a *App) HandlePackages(c *fiber.Ctx) {
	params := initParams(c)

	f, err := remote.Download(params)
	if err != nil {
		intErr(c, err)
	}
	defer os.Remove(params.TempArchive)
	defer f.Close()

	unzipRes, err := zip.Unzip(params.TempArchive, params.TempFolder)
	defer os.RemoveAll(params.TempFolder)
	if err != nil {
		intErr(c, err)
		return
	}

	pkgs, usgs, err := inspectRepo(unzipRes.Module, unzipRes.Root)
	if err != nil {
		intErr(c, err)
		return
	}

	err = a.store.SavePkgs(pkgs)
	if err != nil {
		intErr(c, err)
		return
	}

	err = a.store.SaveUsages(usgs)
	if err != nil {
		intErr(c, err)
		return
	}

	results, err := a.store.GetCombinedUsages()
	if err != nil {
		intErr(c, err)
		return
	}
	c.JSON(results)
}

func initParams(c *fiber.Ctx) model.RepoParams {
	usr := c.Params("user")
	repo := c.Params("repo")
	tmpPath := cfg.Tmp + "/" + strconv.Itoa(int(time.Now().UnixNano()))
	return model.RepoParams{
		GHUser:      usr,
		GHRepo:      repo,
		GHURL:       fmt.Sprintf(cfg.GithubURLTemplate, usr, repo),
		TempFolder:  tmpPath,
		TempArchive: tmpPath + ".zip",
	}
}

func inspectRepo(moduleName, path string) ([]model.Pkg, []model.PkgImportUsage, error) {
	packages := map[string]model.Pkg{}
	usages := []model.PkgImportUsage{}
	err := filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if i.IsDir() {
			return nil
		}
		scanRes, err := goscan.Scan(moduleName, p, i, path)
		if err != nil {
			return err
		}

		pkgPath := fs.LastPartExclude(scanRes.Path)
		existingPkg, ok := packages[pkgPath]
		if !ok {
			existingPkg = model.Pkg{Module: moduleName, Name: scanRes.PkgName, Path: pkgPath}
		}
		existingPkg.Files = append(existingPkg.Files, scanRes.PkgFile)
		packages[pkgPath] = existingPkg
		usages = append(usages, scanRes.ImportUsages...)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	pSlice := make([]model.Pkg, 0, len(packages))
	for _, v := range packages {
		pSlice = append(pSlice, v)
	}
	return pSlice, usages, nil
}

func intErr(c *fiber.Ctx, e error) {
	c.JSON(model.HTTPError{e.Error()})
	c.Status(http.StatusInternalServerError)
}

func err(c *fiber.Ctx, code int, e error) {
	c.JSON(model.HTTPError{e.Error()})
	c.Status(http.StatusInternalServerError)
}

func success(c *fiber.Ctx, data interface{}) {
	c.JSON(data)
	c.Status(http.StatusOK)
}
