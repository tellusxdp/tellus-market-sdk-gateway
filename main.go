package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tellusxdp/tellus-market-sdk-gateway/config"
	"github.com/tellusxdp/tellus-market-sdk-gateway/server"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Action = serve

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Value: "config.yml",
			Usage: "config file",
		},
	}

	log.SetLevel(log.DebugLevel)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func serve(c *cli.Context) error {
	configPath := c.String("config")
	cfg, err := config.FromFilepath(configPath)
	if err != nil {
		return err
	}

	s, err := server.New(cfg)
	if err != nil {
		return err
	}

	err = s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
