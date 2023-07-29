#!/bin/bash

set -x

: ${CONFIG_PATH:=$1}

./sprinkler database initialize --config $CONFIG_PATH
./sprinkler service control --config $CONFIG_PATH
