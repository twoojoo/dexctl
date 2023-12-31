package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/dexidp/dex/api/v2"
	"github.com/twoojoo/dexctl/setup"
	"github.com/twoojoo/dexctl/utils"
	"github.com/urfave/cli"
)

var CreateClientFlags []cli.Flag = []cli.Flag{
	cli.StringFlag{
		Name:  "grpc-url, g",
		Value: "127.0.0.1:5557",
		Usage: "gRPC host and port",
	},
	cli.StringFlag{
		Name:  "id, i",
		Value: "random UUID",
	},
	cli.StringFlag{
		Name:     "name, n",
		Required: true,
	},
	cli.StringFlag{
		Name:  "secret, s",
		Value: utils.RandStringFlag,
	},
	cli.StringSliceFlag{
		Name:  "redirect-uri, r",
		Value: &cli.StringSlice{"http://127.0.0.1:5555/callback"},
	},
	cli.StringFlag{
		Name:     "logo-url, l",
		Required: false,
	},
	cli.StringSliceFlag{
		Name:     "trusted-peer, t",
		Required: false,
	},
	cli.BoolFlag{
		Name:     "public, p",
		Required: false,
	},
}

func CreateClient(c *cli.Context) error {
	ctx := context.Background()

	id, err := utils.ParseRandomUUID(c.String("id"))
	if err != nil {
		return err
	}

	clientSecret := utils.ParseRandomString(c.String("secret"), 40)

	grpc, err := setup.SetupGrpcClient(ctx, c)
	if err != nil {
		return err
	}

	client := &api.Client{
		Id:           id,
		Secret:       clientSecret,
		Name:         c.String("name"),
		RedirectUris: c.StringSlice("redirect-uri"),
		LogoUrl:      c.String("logo-url"),
		TrustedPeers: c.StringSlice("trusted-peer"),
		Public:       c.Bool("public"),
	}

	resp, err := grpc.CreateClient(ctx, &api.CreateClientReq{Client: client})
	if err != nil {
		return err
	}

	if resp.AlreadyExists {
		return errors.New("client already exists")
	}

	p, err := utils.PrettifyJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(p)

	return nil
}
