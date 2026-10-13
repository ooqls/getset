package app

import (
	"flag"
	"strings"
)

var (
	registryPathFlag   string
	sqlFilesFlag       string
	rsaPrivKeyPathFlag string
	rsaPubKeyPathFlag  string
	jwtPrivKeyPathFlag string
	jwtPubKeyPathFlag  string
	tlsKeyPathFlag     string
	tlsCertPathFlag    string
	tlsCaPathFlag      string
	healthPathFlag     string
	docsPathFlag       string
	docsApiPathFlag    string
)

func init() {
	flag.StringVar(&registryPathFlag, "registry", "", "Path to the registry path")
	flag.StringVar(&sqlFilesFlag, "sql-files", "", "Comma separated list of files")
	flag.StringVar(&rsaPrivKeyPathFlag, "rsa-private-key", "", "Path to an RSA private key")
	flag.StringVar(&rsaPubKeyPathFlag, "rsa-public-key", "", "Path to the RSA public key")
	flag.StringVar(&jwtPrivKeyPathFlag, "jwt-private-key", "", "Path to a JWT private key")
	flag.StringVar(&jwtPubKeyPathFlag, "jwt-public-key", "", "Path to a jwt public key")
	flag.StringVar(&tlsKeyPathFlag, "tls-key-path", "", "Path to the TLS key file")
	flag.StringVar(&tlsCertPathFlag, "tls-cert-path", "", "Path to the TLS cert file")
	flag.StringVar(&tlsCaPathFlag, "tls-ca-path", "", "Path to the TLS CA file")
	flag.StringVar(&healthPathFlag, "health-path", "", "Path to the health path")
	flag.StringVar(&docsPathFlag, "docs-path", "", "Path to the docs directory")
	flag.StringVar(&docsApiPathFlag, "docs-api-path", "/api/docs", "Path to the docs API")
}

// applyFlags fills any feature field left empty by options or config from
// the matching command-line flag. Options and config always win.
func applyFlags(fs *Features) {
	setIfEmpty(&fs.Registry.Path, registryPathFlag)
	setIfEmpty(&fs.RSA.PrivateKeyPath, rsaPrivKeyPathFlag)
	setIfEmpty(&fs.RSA.PublicKeyPath, rsaPubKeyPathFlag)
	setIfEmpty(&fs.JWT.PrivateKeyPath, jwtPrivKeyPathFlag)
	setIfEmpty(&fs.JWT.PubKeyPath, jwtPubKeyPathFlag)
	setIfEmpty(&fs.TLS.ServerKeyFile, tlsKeyPathFlag)
	setIfEmpty(&fs.TLS.ServerCertFile, tlsCertPathFlag)
	setIfEmpty(&fs.TLS.CAFile, tlsCaPathFlag)
	setIfEmpty(&fs.Health.Path, healthPathFlag)
	setIfEmpty(&fs.Docs.DocsPath, docsPathFlag)
	setIfEmpty(&fs.Docs.DocsApiPath, docsApiPathFlag)
	if len(fs.SQL.SQLFiles) == 0 && sqlFilesFlag != "" {
		fs.SQL.SQLFiles = strings.Split(sqlFilesFlag, ",")
	}
}

func setIfEmpty(field *string, value string) {
	if *field == "" {
		*field = value
	}
}
