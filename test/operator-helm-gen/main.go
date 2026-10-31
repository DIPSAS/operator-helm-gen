package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DIPSAS/operator-helm-gen/pkg/helmgen"
)

func main() {
	filename := filepath.Join("testdata", "rbac.yaml")
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	dir := "output"

	if err = os.RemoveAll(dir); errors.Is(err, &os.PathError{}) {
		panic(fmt.Errorf("could not clean output dir %s: %w", dir, err))
	}

	if _, err := os.Stat("dir"); os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			panic(fmt.Errorf("could not create output dir %s: %w", dir, err))
		}
	}

	resources := helmgen.GetResources(data)

	annotationsArg := "argoproj.io/sync-wave=123,test/test=test"
	annotations, err := helmgen.GetAnnotationsFromArg(annotationsArg)
	if err != nil {
		panic(fmt.Errorf("could not parse annotation arg: %w ", err))
	}

	whichResources := "RoleBinding,Role"
	spec := helmgen.Spec{KeepResources: helmgen.GetResourcesToKeep(whichResources), OutputDir: dir, Annotations: annotations}

	p := helmgen.NewPatcher(spec)
	err = p.Generate(resources)

	if err != nil {
		panic(fmt.Errorf("could not generate for resource found in file %s: %w", filename, err))
	}
}
