package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/DIPSAS/operator-helm-gen/pkg/helmgen"
)

func main() {
	println(time.Now().UTC().Format(time.RFC3339))
	var keep string
	var dir string
	flag.StringVar(&keep, "keep", "", "Specify which resources to keep, separated by comma.")
	flag.StringVar(&dir, "dir", "", "Specify output dir")
	flag.Parse()

	fmt.Printf("Generating for resources: %s, output dir %s\n", keep, dir)

	var data []byte
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	parts := strings.Split(string(data), "---")
	spec := helmgen.Spec{OutputDir: dir}

	k := make(map[string]string)
	for _, x := range strings.Split(keep, ",") {
		k[x] = "keep"
	}

	spec.KeepRresources = k

	p := &helmgen.Patcher{Spec: spec}
	err = p.Generate(parts)
	if err != nil {
		panic(fmt.Errorf("failed generating files from input: %v", err))
	}
}
