#!/bin/bash - 

if [[ $WITH_DELVE ]]; then
  dlv --listen=:2346 --headless=true --api-version=2 --accept-multiclient debug --build-flags="-gcflags 'all=-N -l'" --continue .
  exit 0
fi

go build -o goapp
./goapp

