package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ooqls/getset/crypto/keys"
	"github.com/urfave/cli"
)

func gen_rsa(ctx *cli.Context) error {
	keyPairName := ctx.String("name")
	rsaOutFlag := ctx.String("out")
	priv, pub, err := keys.NewRsaKeyPemBytes()
	if err != nil {
		return err
	}

	privPath := path.Join(rsaOutFlag, fmt.Sprintf("%s.pem", keyPairName))
	pubPath := path.Join(rsaOutFlag, fmt.Sprintf("%s_pub.pem", keyPairName))
	err = os.WriteFile(privPath, priv, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(pubPath, pub, 0644)
	if err != nil {
		return err
	}

	return nil
}
