package preset

type OutputMode string

const (
	OutputConcat    OutputMode = "concat"
	OutputCopy      OutputMode = "copy"
	OutputTemplate  OutputMode = "template"
	OutputSymlink   OutputMode = "symlink"
	OutputMkdir     OutputMode = "mkdir"
)

type Preset struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	OS          []string          `yaml:"os"`
	Arch        []string          `yaml:"arch"`
	Tags        []string          `yaml:"tags"`
	Required    bool              `yaml:"required"`
	Conflicts   []string          `yaml:"conflicts"`
	Depends     []string          `yaml:"depends"`
	Recommends  []string          `yaml:"recommends"`
	Packages    []string          `yaml:"packages"`
	Outputs     []Output          `yaml:"outputs"`
	Hooks       Hooks             `yaml:"hooks"`
	Config      map[string]string `yaml:"config"`
}

type Output struct {
	Target   string     `yaml:"target"`
	Source   string     `yaml:"source"`
	Priority int        `yaml:"priority"`
	Local    bool       `yaml:"local"`
	Mode     OutputMode `yaml:"mode"`
}

type Hooks struct {
	PreInstall  string `yaml:"pre_install"`
	PostInstall string `yaml:"post_install"`
}

type TemplateData struct {
	Config   map[string]string
	OS       string
	Arch     string
	Home     string
	XDG      string
	Hostname string
	User     string
	OHP1X    string
}
