#!/bin/bash
# encoding:utf-8
#Author:sumeru
#Date:2019-09-13
#Script:trpc_monitor.sh
#Description:watch the process info

#get trpc framework config
bashPath=/usr/local/app
trpcPath=/usr/local/trpc/bin
logFile=$bashPath/monitor.log
processNumErr=10002
pidHasChanged=10003
processNotExist=10004

#check process num
num=`ps -ef |grep $trpcPath/$SUMERU_SERVER |grep -v grep|wc -l`
#num=`ps -ef |grep $trpcPath/helloworld |grep -v grep|wc -l`
echo "`date` the num of process is $num">>$logFile
if [ $num -lt 1 ];then
    echo "`date` the process is not exist now and begin check start.sh">>$logFile
        startSriptNum=`ps -ef |grep $bashPath/start.sh|grep -v grep|wc -l`
        if [ $startSriptNum -lt 1 ];then
                nohup bash $bashPath/start.sh 2>&1 >>$logFile 2>&1 &
                echo "`date` the start.sh is not executing now so begin execute start.sh in background ">>$logFile
        else
                echo "`date` the start.sh is executing now so do noting ">>$logFile
        fi
    exit $processNotExist
fi

#if pid.conf is exist,the pid check is needed
if [ -f $bashPath/pid.conf ];then
        pidNow=`ps -ef |grep $trpcPath/$SUMERU_SERVER |grep -v grep|head -n 1 |awk '{print $2}'`
        pidOld=`cat $bashPath/pid.conf`
        if [ $pidNow -ne $pidOld ];then
                echo "`date` the pid now $pidNow is not equel with pidInfo in pid.conf $pidOld">>$logFile
                exit $pidHasChanged
        fi
fi

exit 0
