#!/usr/bin/env bash
u=$1
shift
p=github.com/$u/mymysql

go $* $p/mysql $p/native $p/thrsafe $p/autorc $p/godrv
