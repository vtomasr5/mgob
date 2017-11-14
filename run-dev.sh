#! /bin/bash

go install

mkdir -p /tmp/mgob/config
mkdir -p /tmp/mgob/data
mkdir -p /tmp/mgob/storage
mkdir -p /tmp/mgob/tmp

cp test/config/mongo-local.yml /tmp/mgob/config/

mgob -ConfigPath /tmp/mgob/config/ -DataPath /tmp/mgob/data/ -StoragePath /tmp/mgob/storage -TmpPath /tmp/mgob/tmp/
