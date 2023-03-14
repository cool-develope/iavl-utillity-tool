
load-snapshot:
	go run ./... load-snapshot -s /.gaia/data/snapshots -a /.gaia/data
.PHONY: load-snapshot


sync-snapshot:
	go run ./... synchronize -s /.gaia/data/snapshots -a /.gaia/data_new