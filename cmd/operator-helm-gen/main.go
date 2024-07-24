package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/DIPSAS/operator-helm-gen/pkg/helmgen"
)

func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "(unknown)"
	}

	return info.Main.Version
}

func main() {
	var keep string
	var dir string
	var showVersion bool
	flag.StringVar(&keep, "keep", "", "Specify which resources to keep, separated by comma.")
	flag.StringVar(&dir, "dir", "", "Specify output dir")
	flag.BoolVar(&showVersion, "version", false, "Prints the version and exits.")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s\n", Version())
		return
	}

	fmt.Printf("Generating for resources: %s, output dir %s\n", keep, dir)

	var data []byte
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	resources := helmgen.GetResources(data)
	spec := helmgen.Spec{OutputDir: dir}

	k := make(map[string]string)
	for _, x := range strings.Split(keep, ",") {
		k[x] = "keep"
	}

	spec.KeepRresources = k

	p := &helmgen.Patcher{Spec: spec}
	err = p.Generate(resources)
	if err != nil {
		panic(fmt.Errorf("failed generating files from input: %v", err))
	}
}
