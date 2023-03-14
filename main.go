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
	flagVersion     = "version"
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
	versionFlag = cli.Int64Flag{
		Name:     flagVersion,
		Aliases:  []string{"v"},
		Usage:    "The block height to load and sync",
		Required: false,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = version

	app.Commands = []*cli.Command{
		{
			Name:    "save-snapshot",
			Aliases: []string{},
			Usage:   "Load the app-state and save the snapshot",
			Action:  saveSnapshot,
			Flags: []cli.Flag{
				&storeDirFlag,
				&snapshotDirFlag,
				&versionFlag,
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
				&versionFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
