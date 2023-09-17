#!/bin/bash

ENDPOINT=$1
ROLE=$2
TOKEN=$3

docker run --rm -t -v $PWD/_out:/out \
  ghcr.io/siderolabs/imager:v1.5.2 iso \
--system-extension-image ghcr.io/siderolabs/gvisor:20231214.0-v1.5.2-1-g3663e39 \
--system-extension-image ghcr.io/siderolabs/intel-ucode:20230613 \
--extra-kernel-arg net.ifnames=0 \
--output /out/${ROLE} \
--extra-kernel-arg talos.config="${ENDPOINT}/register?role=${ROLE}&token=${TOKEN}"
