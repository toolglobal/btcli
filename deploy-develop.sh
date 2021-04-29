#!/bin/bash

# btcli
cd /root/apps/btcli/
chmod +x btcli
ps -ef | grep -w btcli | grep -v grep | awk  '{print "kill -9 " $2}' | sh
sleep 1
nohup ./btcli 2>&1 >> ./output.log 2>&1 &