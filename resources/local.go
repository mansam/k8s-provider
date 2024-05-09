package resources

import (
	"errors"
	"io"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

func decodeFile(path string) (resources []any, err error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return
	}
	decoder := yaml.NewDecoder(f)
	for {
		var resource any
		err = decoder.Decode(&resource)
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		f.Name()
		resources = append(resources, resource)
	}
	return
}

func LoadLocalResources(location string) (resources []any, err error) {
	entries, err := os.ReadDir(location)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			r, rErr := LoadLocalResources(path.Join(location, entry.Name()))
			if rErr != nil {
				err = rErr
				return
			}
			resources = append(resources, r...)
		} else {
			r, rErr := decodeFile(path.Join(location, entry.Name()))
			if rErr != nil {
				err = rErr
				return
			}
			resources = append(resources, r...)
		}
	}
	return
}
