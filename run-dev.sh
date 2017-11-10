#! /bin/bash
#
# run-test.sh
# Copyright (C) 2017 vjuan <vtomasr5@gmail.com>
#
# Distributed under terms of the GPLv3 license.

go install

mkdir -p /tmp/config
mkdir -p /tmp/data
mkdir -p /tmp/storage
mkdir -p /tmp/tmp

cp test/config/mongo-dev.yml /tmp/config/

mgob -ConfigPath /tmp/config/ -DataPath /tmp/data/ -StoragePath /tmp/storage -TmpPath /tmp/tmp/
