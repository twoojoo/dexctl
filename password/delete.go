package password

import (
	"context"
	"errors"
	"fmt"

	"github.com/dexidp/dex/api/v2"
	"github.com/twoojoo/dexctl/setup"
	"github.com/twoojoo/dexctl/utils"
	"github.com/urfave/cli"
)

var DeletePasswordFlags []cli.Flag = []cli.Flag{
	cli.StringFlag{
		Name:  "grpc-url, g",
		Value: "127.0.0.1:5557",
		Usage: "gRPC host and port",
	},
}

func DeletePassword(c *cli.Context) error {
	ctx := context.Background()

	grpc, err := setup.SetupGrpcClient(ctx, c)
	if err != nil {
		return err
	}

	email, err := utils.ParseEmail(c.Args().Get(0))
	if err != nil {
		return err
	}

	resp, err := grpc.DeletePassword(ctx, &api.DeletePasswordReq{
		Email: email,
	})

	if err != nil {
		return err
	}

	if resp.NotFound {
		return errors.New("password not found")
	}

	p, err := utils.PrettifyJSON(resp)
	if err != nil {
		return err
	}

	fmt.Println(p)

	return nil
}
