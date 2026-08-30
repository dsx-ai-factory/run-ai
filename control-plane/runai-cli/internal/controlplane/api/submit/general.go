package submit

type General struct {
	PrefixName      string
	UseNameAsPrefix bool
	Labels          []string
	Annotations     []string
	DryRun          bool
	Template        string
}
