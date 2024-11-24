#!/bin/bash
# encoding:utf-8
#Author:sumeru
#Date:2019-09-13
#Script:trpc_stop.sh
#Description:stop process

#try to stop process in elegant way
bashPath=/usr/local/app
trpcPath=/usr/local/trpc/bin
logFile=$bashPath/stop.log
processNotExist=10005

#stop process
pidInfo=$(ps -ef | grep "$trpcPath/$SUMERU_SERVER" | grep -v grep | awk '{print $2}')

echo "`date` the pid info is $pidInfo">>$logFile

for pid in $pidInfo;do
        kill -9 $pid
done

sleep 3

#check process num
num=$(ps -ef | grep "$trpcPath/$SUMERU_SERVER" | grep -v grep |wc -l)

if [ $num -gt 0 ];then
        echo "`date` after stop process in force way the processNum is $num still bigger than 0">>$logFile
        exit $processNotExist
fi
exit 0