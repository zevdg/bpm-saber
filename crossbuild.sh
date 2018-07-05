#!/bin/bash

docker run -u $(id -u):$(id -g) -e XDG_CACHE_HOME='/tmp/.cache' -v $(pwd):/go/src/meter-saber zevdg/go-ui-crossbuild gouicrossbuild meter-saber . release

mv release/meter-saber release/meter-saber_linux
zip release/meter-saber_mac.zip release/meter-saber.app && rm release/meter-saber.app