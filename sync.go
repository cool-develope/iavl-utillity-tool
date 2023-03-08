package main

import (
	"fmt"
	"os"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/urfave/cli/v2"
)

func synchronize(ctx *cli.Context) error {
	logger := log.NewLogger()
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	snapshotDir := dirname + "/" + ctx.String(flagSnapshotDir) // "/.simapp/data/snapshots"

	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		return err
	}
	defer snapshotDB.Close()

	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		return err
	}
	snapshot, err := snapshotStore.GetLatest()
	if err != nil {
		return err
	}

	_, chunks, err := snapshotStore.Load(snapshot.Height, snapshot.Format)
	if err != nil {
		return err
	}
	streamReader, err := snapshots.NewStreamReader(chunks)
	if err != nil {
		return err
	}
	defer streamReader.Close()

	appDir := dirname + ctx.String(flagStoreDir) // "/.simapp/data"
	appStoreDB, err := dbm.NewDB("application", dbm.GoLevelDBBackend, appDir)
	if err != nil {
		return err
	}
	defer appStoreDB.Close()

	keys := storetypes.NewKVStoreKeys(
		"acc", "bank", "staking", "slashing", "gov", "upgrade", "mint", "distribution", "consensus",
	)
	mstore := store.NewCommitMultiStore(appStoreDB, logger, nil)
	for _, key := range keys {
		mstore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}
	if err = mstore.LoadLatestVersion(); err != nil {
		return err
	}

	nextItem, err := mstore.Restore(snapshot.Height, snapshot.Format, streamReader)
	fmt.Printf("nextItem: %v\n", nextItem)

	return err
}
