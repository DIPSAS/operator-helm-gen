package helmgen

type Spec struct {
	//which resources to work on.
	KeepResources map[string]string
	//the output directory.
	OutputDir string
	//annotations to insert on the generated results.
	Annotations map[string]string
}
