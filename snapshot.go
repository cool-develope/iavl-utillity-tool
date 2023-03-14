package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	"cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	"github.com/golang/snappy"
	"github.com/urfave/cli/v2"
)

var storeKeys = storetypes.NewKVStoreKeys(
	"acc", "bank", "staking", "slashing", "gov", "upgrade", "mint",
)

func saveSnapshot(ctx *cli.Context) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	appStoreDir := dirname + ctx.String(flagStoreDir)    // "/.simapp/data"
	snapshotDir := dirname + ctx.String(flagSnapshotDir) // "/.simapp/data/snapshots"
	version := ctx.Int64(flagVersion)

	mstore, err := loadAppStore(appStoreDir, version)
	if err != nil {
		return err
	}

	if err = createSnapshot(snapshotDir, mstore, version); err != nil {
		return err
	}

	return err
}

func loadAppStore(appStoreDir string, version int64) (store.CommitMultiStore, error) {
	appStoreDB, err := dbm.NewDB("application", dbm.GoLevelDBBackend, appStoreDir)
	if err != nil {
		return nil, err
	}

	logger := log.NewNopLogger()
	mstore := store.NewCommitMultiStore(appStoreDB, logger, nil)
	for _, key := range storeKeys {
		mstore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}

	if err = mstore.LoadLatestVersion(); err != nil {
		return nil, err
	}
	latestVersion := mstore.LatestVersion()

	if version != 0 {
		if err = mstore.LoadVersion(version); err != nil {
			return nil, err
		}

		if err = writeChangeSets(appStoreDir, appStoreDB, version, latestVersion); err != nil {
			return nil, err
		}
	}

	return mstore, nil
}

func createSnapshot(snapshotDir string, mstore store.CommitMultiStore, version int64) error {
	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		return err
	}
	defer snapshotDB.Close()

	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		return err
	}

	m := snapshots.NewManager(snapshotStore, types.SnapshotOptions{}, mstore, nil, nil)
	if version == 0 {
		version = mstore.LatestVersion()
	}
	if _, err = m.Prune(0); err != nil {
		return err
	}
	_, err = m.Create(uint64(version))
	return err
}

func writeChangeSets(appStoreDir string, appStoreDB dbm.DB, startVersion, endVersion int64) error {
	for _, storeKey := range storeKeys {
		snapshotFile := filepath.Join(appStoreDir, "snapshots", fmt.Sprintf("%s-%d-%d.snappy", storeKey.Name(), startVersion, endVersion))
		fp, err := os.OpenFile(snapshotFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return err
		}
		writer := snappy.NewBufferedWriter(fp)
		prefix := []byte(fmt.Sprintf("s/k:%s/", storeKey.Name()))
		tree, err := iavl.NewMutableTree(dbm.NewPrefixDB(appStoreDB, prefix), 0, true)
		if err != nil {
			return err
		}
		if err := tree.TraverseStateChanges(startVersion, endVersion, func(version int64, changeSet *iavl.ChangeSet) error {
			return WriteChangeSet(writer, version, *changeSet)
		}); err != nil {
			return err
		}

		writer.Flush()
	}
	return nil
}
