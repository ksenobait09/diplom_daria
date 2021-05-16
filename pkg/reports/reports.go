package reports

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Repo struct {
	Directory string
}

type Report struct {
	Name string
	Href string
}

func New(dir string) *Repo {
	return &Repo{Directory: dir}
}

func (r *Repo) List() ([]Report, error) {
	files, err := ioutil.ReadDir(r.Directory)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", r.Directory)
	}

	res := make([]Report, 0, len(files))
	for _, file := range files {
		fileName := file.Name()
		res = append(res, Report{
			Name: fileName[:strings.LastIndex(fileName, ".")],
			Href: fileName})
	}

	return res, nil
}

func (r *Repo) Add(reportName string, data io.Reader) error {
	filepath := r.Directory + "/" + reportName

	body, err := ioutil.ReadAll(data)
	if err != nil {
		return errors.Wrap(err, "failed to read file from reader")
	}

	err = ioutil.WriteFile(filepath, body, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write file %s", filepath)
	}
	return nil
}

func (r *Repo) Delete(reportName string) error {
	filepath := r.Directory + "/" + reportName
	err := os.Remove(filepath)
	if err != nil {
		return errors.Wrapf(err, "failed to remove file %s", filepath)
	}

	return nil
}
