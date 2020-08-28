package remote

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/trapck/gopref/model"
)

// Download downloads github repo as zip archive
func Download(params model.RepoParams) (*os.File, error) {
	resp, err := http.Get(params.GHURL)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d %s", resp.StatusCode, string(b))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(params.TempArchive)
	if err != nil {
		return nil, err
	}
	f.Write(b)
	return f, nil
}
