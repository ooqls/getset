package main

import (
	"fmt"
	"os"
	"path"

	"github.com/ooqls/getset/crypto/keys"
	"github.com/urfave/cli"
)

func genX509Cert(c *cli.Context) error {
	keypair := c.String("name")
	out := c.String("out")
	commonName := c.String("common-name")
	aliases := c.StringSlice("aliases")
	ca := c.String("ca")
	var err error
	var caKeys *keys.X509
	if ca == "" {
		caKeys, err = keys.CreateX509CA()
	} else {
		caKeys, err = keys.ParseX509File(ca)
	}
	if err != nil {
		return err
	}

	keys, err := keys.CreateX509(*caKeys, keys.WithCommonName(commonName), keys.WithDNSNames(aliases))
	if err != nil {
		return err
	}

	priv, pub := keys.Pem()

	privPath := path.Join(out, fmt.Sprintf("%s.pem", keypair))
	pubPath := path.Join(out, fmt.Sprintf("%s_pub.pem", keypair))

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

func genX509CA(ctx *cli.Context) error {
	aliases := ctx.StringSlice("aliases")
	cn := ctx.String("common-name")
	out := ctx.String("out")
	keyPairName := ctx.String("name")

	ca, err := keys.CreateX509CA(
		keys.WithCommonName(cn),
		keys.WithDNSNames(aliases),
	)
	if err != nil {
		return err
	}

	privKeyOut := path.Join(out, fmt.Sprintf("ca_%s.pem", keyPairName))
	pubKeyOut := path.Join(out, fmt.Sprintf("ca_%s_pub.pem", keyPairName))

	priv, pub := ca.Pem()

	err = os.WriteFile(privKeyOut, priv, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(pubKeyOut, pub, 0644)
	if err != nil {
		return err
	}

	return nil

}
