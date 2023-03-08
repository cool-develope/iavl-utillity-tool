package main

import (
	"os"

	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	"cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/urfave/cli/v2"
)

func loadSnapshot(ctx *cli.Context) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	snapshotDir := dirname + ctx.String(flagSnapshotDir) // "/.simapp/data/snapshots"

	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		return err
	}
	defer snapshotDB.Close()

	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		return err
	}

	appDir := dirname + ctx.String(flagStoreDir) // "/.simapp/data"
	appStoreDB, err := dbm.NewDB("application", dbm.GoLevelDBBackend, appDir)
	if err != nil {
		return err
	}
	defer appStoreDB.Close()

	keys := storetypes.NewKVStoreKeys(
		"acc", "bank", "staking", "slashing", "gov", "upgrade", "mint", "distribution", "consensus",
	)
	mstore := store.NewCommitMultiStore(appStoreDB, nil, nil)
	for _, key := range keys {
		mstore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}
	if err = mstore.LoadLatestVersion(); err != nil {
		return err
	}
	m := snapshots.NewManager(snapshotStore, types.SnapshotOptions{}, mstore, nil, nil)
	_, err = m.Create(uint64(mstore.LatestVersion()))

	return err
}
