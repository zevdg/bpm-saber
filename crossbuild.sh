#!/bin/bash

docker run -u $(id -u):$(id -g) -e XDG_CACHE_HOME='/tmp/.cache' -v $(pwd):/go/src/meter-saber zevdg/go-ui-crossbuild gouicrossbuild meter-saber . release