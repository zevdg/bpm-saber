#!/bin/bash

docker run -u $(id -u):$(id -g) -e XDG_CACHE_HOME='/tmp/.cache' -v $(pwd):/go/src/bpm-saber zevdg/go-ui-crossbuild gouicrossbuild bpm-saber . release

mv release/bpm-saber release/bpm-saber_linux
zip release/bpm-saber_mac.zip release/bpm-saber.app && rm release/bpm-saber.app