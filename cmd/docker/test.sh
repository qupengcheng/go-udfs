#!/bin/bash
set -e

cd ../ipfs
GOOS=linux GOARCH=amd64 go build -o ../docker/ipfs_linux

cd ../docker


# FILENAME=./runipfs.sh
# scp ${FILENAME} ipfs-192.168.12.221:/root/

# scp ${FILENAME} ipfs-192.168.12.223:/root/

# scp ${FILENAME} ipfs-192.168.12.224:/root/

# scp ${FILENAME} ipfs-192.168.12.225:/root/

# scp ${FILENAME} ipfs-192.168.12.226:/root/

# scp ${FILENAME} ipfs-192.168.12.227:/root/

IPFS=./ipfs_linux
scp ${IPFS} ipfs-192.168.12.221:/usr/local/bin/ipfs

scp ${IPFS} ipfs-192.168.12.223:/usr/local/bin/ipfs

scp ${IPFS} ipfs-192.168.12.224:/usr/local/bin/ipfs

scp ${IPFS} ipfs-192.168.12.225:/usr/local/bin/ipfs

scp ${IPFS} ipfs-192.168.12.226:/usr/local/bin/ipfs

scp ${IPFS} ipfs-192.168.12.227:/usr/local/bin/ipfs