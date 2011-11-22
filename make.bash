#!/usr/bin/env bash

for dir in mysql native thrsafe autorc godrv examples; do
	(cd $dir; make $@)
done
