package main

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/iavl"
	"cosmossdk.io/store/snapshots"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	iavltree "github.com/cosmos/iavl"
	"github.com/urfave/cli/v2"
)

func synchronize(ctx *cli.Context) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	snapshotDir := dirname + "/" + ctx.String(flagSnapshotDir) // "/.simapp/data/snapshots"
	appDir := dirname + ctx.String(flagStoreDir)               // "/.simapp/data"
	version := ctx.Int64(flagVersion)

	mstore, err := restoreSnapshot(appDir, snapshotDir)
	if err != nil {
		return err
	}

	startVersion := mstore.LatestVersion()
	if version > startVersion {
		suffix := fmt.Sprintf("%d-%d", startVersion+1, version)
		if err = restoreChangeSet(mstore, snapshotDir, suffix); err != nil {
			return err
		}
	}

	return nil
}

func restoreSnapshot(appStoreDir, snapshotDir string) (storetypes.CommitMultiStore, error) {
	logger := log.NewNopLogger()

	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		return nil, err
	}
	defer snapshotDB.Close()

	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		return nil, err
	}
	snapshot, err := snapshotStore.GetLatest()
	if err != nil {
		return nil, err
	}

	_, chunks, err := snapshotStore.Load(snapshot.Height, snapshot.Format)
	if err != nil {
		return nil, err
	}
	streamReader, err := snapshots.NewStreamReader(chunks)
	if err != nil {
		return nil, err
	}
	defer streamReader.Close()

	appStoreDB, err := dbm.NewDB("application", dbm.GoLevelDBBackend, appStoreDir)
	if err != nil {
		return nil, err
	}
	defer appStoreDB.Close()

	mstore := store.NewCommitMultiStore(appStoreDB, logger, nil)
	for _, key := range storeKeys {
		mstore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}
	if err = mstore.LoadLatestVersion(); err != nil {
		return nil, err
	}

	nextItem, err := mstore.Restore(snapshot.Height, snapshot.Format, streamReader)

	fmt.Printf("nextItem: %v\n", nextItem)

	return mstore, err
}

func restoreChangeSet(mstore storetypes.CommitMultiStore, snapshotDir string, suffix string) error {
	for _, storeKey := range storeKeys {
		iavlstore := mstore.GetCommitKVStore(storeKey).(*iavl.Store)
		changesetFile := filepath.Join(snapshotDir, fmt.Sprintf("%s-%s.snappy", storeKey.Name(), suffix))
		fp, err := openChangeSetFile(changesetFile)
		if err != nil {
			return err
		}
		IterateChangeSets(fp, func(version int64, cs *iavltree.ChangeSet) (bool, error) {
			for _, pair := range cs.Pairs {
				if pair.Delete {
					iavlstore.Delete(pair.Key)
				} else {
					iavlstore.Set(pair.Key, pair.Value)
				}
			}
			return false, nil
		})
	}
	return nil
}
