package tar

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Download(url string, path string) error {
	raw, err := http.Get(url)
	if err != nil {
		return NewTarError(err)
	}
	if raw.StatusCode != 200 {
		return NewUnexpectedStatusCodeError(raw.StatusCode)
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return NewTarError(err)
	}
	gzr, err := gzip.NewReader(raw.Body)
	if err != nil {
		return NewTarError(err)
	}
	tr := tar.NewReader(gzr)
	for h, err := tr.Next(); err == nil; h, err = tr.Next() {
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(fmt.Sprintf("%s/%s", path, h.Name), os.ModePerm); err != nil {
				if os.IsExist(err) {
					return nil
				}
				return NewTarError(err)
			}
		case tar.TypeReg:
			file, err := os.Create(fmt.Sprintf("%s/%s", path, h.Name))
			if err != nil {
				return NewTarError(err)
			}
			defer file.Close()
			if _, err := io.Copy(file, tr); err != nil {
				return NewTarError(err)
			}
		}
	}
	if errors.Is(err, io.EOF) {
		return NewTarError(err)
	}
	return nil
}

func DownloadGz(url string, path string) error {
	raw, err := http.Get(url)
	if err != nil {
		return NewTarError(err)
	}
	if raw.StatusCode != 200 {
		return NewUnexpectedStatusCodeError(raw.StatusCode)
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return NewTarError(err)
	}
	gzr, err := gzip.NewReader(raw.Body)
	if err != nil {
		return NewTarError(err)
	}
	tr := tar.NewReader(gzr)
	for h, err := tr.Next(); err == nil; h, err = tr.Next() {
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(fmt.Sprintf("%s/%s", path, h.Name), os.ModePerm); err != nil {
				if os.IsExist(err) {
					continue
				}
				return NewTarError(err)
			}
		case tar.TypeReg:
			file, err := os.Create(fmt.Sprintf("%s/%s", path, h.Name))
			if err != nil {
				return NewTarError(err)
			}
			defer file.Close()
			if _, err := io.Copy(file, tr); err != nil {
				return NewTarError(err)
			}
		}
	}
	if errors.Is(err, io.EOF) {
		return NewTarError(err)
	}
	return nil
}
