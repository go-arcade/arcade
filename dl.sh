#!/bin/bash

target_dir="./internal/engine/router/static"

if [ ! -d $target_dir ]; then
    mkdir -p $target_dir
fi

tag=$(curl -sX GET https://api.github.com/repos/go-arcade/docs/releases/latest | awk '/tag_name/{print $4;exit}' FS='[""]')

if ! curl -o dist.zip -L "https://github.com/go-arcade/docs/releases/download/${tag}/dist.zip"; then
    echo "Failed to download ${tag} dist.zip"
    exit 1
fi

if ! unzip -o -d "$target_dir" dist.zip; then
    echo "Failed to unzip ${tag} dist.zip into ${target_dir}"
    exit 2
fi

rm -f dist.zip
