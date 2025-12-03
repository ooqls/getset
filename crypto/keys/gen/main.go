package main

import (
	"os"

	"github.com/urfave/cli"
)

var aliasesFlag cli.StringSliceFlag = cli.StringSliceFlag{
	Name:  "aliases",
	Usage: "other aliases valid for this cert",
}

var commonNameFlag cli.StringFlag = cli.StringFlag{
	Name:  "common-name",
	Usage: "The common name valid for this cert",
}

var outFlag cli.StringFlag = cli.StringFlag{
	Name:  "out",
	Usage: "the folder to place the keypair",
	Value: "./",
}

var keypairFlag cli.StringFlag = cli.StringFlag{
	Name:  "name",
	Usage: "name for the key pair",
	Value: "x509",
}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:      "rsa",
			ShortName: "r",
			Usage:     "Generate RSA private/public key pairs",
			Flags: []cli.Flag{
				keypairFlag,
				outFlag,
			},
			Action: gen_rsa,
		},
		{
			Name: "x509",
			Subcommands: cli.Commands{
				{
					Name:      "cert",
					ShortName: "crt",
					Usage:     "Generates an x509 cert signed by the given CA",
					Action:    genX509Cert,
					Flags: []cli.Flag{
						outFlag,
						keypairFlag,
						commonNameFlag,
						aliasesFlag,
						cli.StringFlag{
							Name:  "ca",
							Usage: "path to the CA private key",
						},
					},
				},
				{
					Name:   "ca",
					Usage:  "Generates an x509 CA",
					Action: genX509CA,
					Flags: []cli.Flag{
						outFlag,
						keypairFlag,
						commonNameFlag,
						aliasesFlag,
					},
				},
			},
		},
	}
	app.Run(os.Args)

}
