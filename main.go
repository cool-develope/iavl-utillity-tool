package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

const (
	appName = "store-analyzer"
	version = "0.0.1"
)

const (
	flagStoreDir    = "app-store-dir"
	flagSnapshotDir = "snapshot-dir"
)

var (
	storeDirFlag = cli.StringFlag{
		Name:     flagStoreDir,
		Aliases:  []string{"a"},
		Usage:    "The directory where the app store is located",
		Required: true,
	}
	snapshotDirFlag = cli.StringFlag{
		Name:     flagSnapshotDir,
		Aliases:  []string{"s"},
		Usage:    "The directory where the snapshot store is located",
		Required: true,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = version

	app.Commands = []*cli.Command{
		{
			Name:    "load-snapshot",
			Aliases: []string{},
			Usage:   "Load the app-state and save the snapshot",
			Action:  loadSnapshot,
			Flags: []cli.Flag{
				&storeDirFlag,
				&snapshotDirFlag,
			},
		},
		{
			Name:    "synchronize",
			Aliases: []string{},
			Usage:   "Synchronize the snapshot",
			Action:  synchronize,
			Flags: []cli.Flag{
				&storeDirFlag,
				&snapshotDirFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
