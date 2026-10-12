package app

import (
	"net/http"
	"os"

	"go.uber.org/zap"
)

type DocsConfig struct {
	Enabled     bool   `yaml:"enabled"`
	DocsApiPath string `yaml:"docs_api_path"`
	DocsDir     string `yaml:"docs_dir"`
	DocsContent string `yaml:"docs_content"`
}

type DocsOption func(*DocsFeature)

func WithDocsPath(path string) DocsOption {
	return func(f *DocsFeature) { f.DocsPath = path }
}

func WithDocsApiPath(path string) DocsOption {
	return func(f *DocsFeature) { f.DocsApiPath = path }
}

func WithDocsContent(content string) DocsOption {
	return func(f *DocsFeature) { f.DocsContent = content }
}

// DocsFeature serves docs at DocsApiPath on the Gin and HTTP routers: either
// DocsContent as a single HTML page, or the files under the DocsPath
// directory.
type DocsFeature struct {
	Enabled     bool
	DocsPath    string
	DocsApiPath string
	DocsContent string
}

func Docs(opts ...DocsOption) DocsFeature {
	f := DocsFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func docsFromConfig(c DocsConfig) DocsFeature {
	return DocsFeature{
		Enabled:     c.Enabled,
		DocsPath:    c.DocsDir,
		DocsApiPath: c.DocsApiPath,
		DocsContent: c.DocsContent,
	}
}

func (f *DocsFeature) name() string  { return "docs" }
func (f *DocsFeature) enabled() bool { return f.Enabled }

func (f *DocsFeature) start(ctx *AppContext, fs *Features) error {
	ctx.L().Info("[Startup docs] serving docs",
		zap.String("path", f.DocsPath), zap.String("api_path", f.DocsApiPath))

	if f.DocsContent != "" {
		content := []byte(f.DocsContent)
		fs.handle(f.DocsApiPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(content)
		}))
	} else if f.DocsPath != "" {
		fs.handlePrefix(f.DocsApiPath, http.FileServer(http.FS(os.DirFS(f.DocsPath))))
	}
	return nil
}
