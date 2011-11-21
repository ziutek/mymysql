#!/usr/bin/env bash

for dir in mysql native thrsafe autorc examples; do
	(cd $dir; make $@)
done
