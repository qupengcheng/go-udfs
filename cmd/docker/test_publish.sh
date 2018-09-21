#!/bin/bash
set -e

EXENAME=ipfs_linux
cd ../ipfs
GOOS=linux GOARCH=amd64 go build -o ../docker/${EXENAME}

cd ../docker

# FILENAME=./runipfs.sh
# scp ${FILENAME} ipfs-192.168.12.221:/root/

# scp ${FILENAME} ipfs-192.168.12.223:/root/

# scp ${FILENAME} ipfs-192.168.12.224:/root/

# scp ${FILENAME} ipfs-192.168.12.225:/root/

# scp ${FILENAME} ipfs-192.168.12.226:/root/

# scp ${FILENAME} ipfs-192.168.12.227:/root/

# scp ${EXENAME} ipfs-192.168.12.221:/usr/local/bin/ipfs

# scp ${EXENAME} ipfs-192.168.12.223:/usr/local/bin/ipfs

# scp ${EXENAME} ipfs-192.168.12.224:/usr/local/bin/ipfs

# scp ${EXENAME} ipfs-192.168.12.225:/usr/local/bin/ipfs

# scp ${EXENAME} ipfs-192.168.12.226:/usr/local/bin/ipfs

# scp ${EXENAME} ipfs-192.168.12.227:/usr/local/bin/ipfs


scp ${EXENAME} ulord-111.231.218.88:/home/ubuntu/bin/ipfs

scp ${EXENAME} ulord-132.232.97.154:/home/ubuntu/bin/ipfs

scp ${EXENAME} ulord-132.232.99.236:/home/ubuntu/bin/ipfs

scp ${EXENAME} ulord-132.232.99.251:/home/ubuntu/bin/ipfs