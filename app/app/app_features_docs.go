package app

var docsPortFlag int
var docsPathFlag string
var docsApiPathFlag string

const (
	docs_pathOpt     string = "opt-docs-path"
	docs_api_pathOpt string = "opt-docs-api-path"
	docs_portOpt     string = "opt-docs-port"
	docs_contentOpt  string = "opt-docs-content"
)

type docsOpt struct {
	featureOpt
}

func WithDocsPath(path string) docsOpt {
	return docsOpt{featureOpt: featureOpt{key: docs_pathOpt, value: path}}
}

func WithDocsApiPath(path string) docsOpt {
	return docsOpt{featureOpt: featureOpt{key: docs_api_pathOpt, value: path}}
}

func WithDocsPort(port int) docsOpt {
	return docsOpt{featureOpt: featureOpt{key: docs_portOpt, value: port}}
}

func WithDocsContent(content string) docsOpt {
	return docsOpt{featureOpt: featureOpt{key: docs_contentOpt, value: content}}
}

type DocsFeature struct {
	Enabled     bool
	DocsPath    string
	DocsApiPath string
	DocsPort    int
	DocsContent string
}

func (f *DocsFeature) apply(opt docsOpt) {
	switch opt.key {
	case docs_pathOpt:
		f.DocsPath = opt.value.(string)
	case docs_api_pathOpt:
		f.DocsApiPath = opt.value.(string)
	case docs_portOpt:
		f.DocsPort = opt.value.(int)
	case docs_contentOpt:
		f.DocsContent = opt.value.(string)
	}
}

func Docs(opts ...docsOpt) DocsFeature {
	f := DocsFeature{
		Enabled:     true,
		DocsPath:    docsPathFlag,
		DocsApiPath: docsApiPathFlag,
		DocsPort:    docsPortFlag,
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}
