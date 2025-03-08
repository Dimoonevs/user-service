#!/usr/bin/env bash

sudo pm2 stop user-service
sudo GOMAXPROCS=3 pm2 start user-service-linux-amd64 --name=user-service -- -config=./prod.ini
sudo pm2 save