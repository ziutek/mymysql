#!/usr/bin/env bash

for dir in mysql native thrsafe autorc; do
	(cd $dir; make $@)
done
