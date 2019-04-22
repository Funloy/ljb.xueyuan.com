#!/bin/sh
USER=$1
GROUP=$2
BASEPATH="$(cd "$(dirname "$0")";pwd)"
echo [Unit] >> maiyajia.service
echo Description=Maiyajia.com Server API >> maiyajia.service
echo After=network.target >> maiyajia.service
echo After=mangodb.service >> maiyajia.service

echo [Service] >> maiyajia.service
echo User=$USER >> maiyajia.service
echo Group=$GROUP >> maiyajia.service
echo ExecStart="$BASEPATH"/maiyajia.com >> maiyajia.service
echo Restart=always >> maiyajia.service
echo WorkingDirectory="$BASEPATH" >> maiyajia.service

echo [Install] >> maiyajia.service
echo WantedBy=multi-user.target >> maiyajia.service
chmod 777 maiyajia.service
sudo mv maiyajia.service /etc/systemd/system/maiyajia.service
sudo systemctl daemon-reload
