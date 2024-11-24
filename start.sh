#!/bin/bash
# encoding:utf-8
# Author: smallhow
# Date:2024-09-13
# Script:start.sh
# Description: start process

#get trpc framework config
bashPath=/usr/local/app
trpcPath=/usr/local/trpc/bin
logFile=$bashPath/startScript.log
serverLogFile=$bashPath/server.log
serverHistoryLogFile=$bashPath/serverHistory.log
headerConfFile=$bashPath/header.conf
trpcConfFile=$bashPath/trpc.conf
pidConf=$trpcPath/pid.conf

#err code
getConfFail=10000
fileMissing=10001
startProcessFail=10002
source /etc/profile

if [ ! -d $bashPath ];then
    mkdir -p $bashPath
fi
mkdir -p /usr/local/trpc/log

#check process
num=`ps -ef |grep $trpcPath/$SUMERU_SERVER |grep -v grep|wc -l`
echo "`date` the num of process is $num">>$logFile
if [ $num -gt 0 ];then
        echo "the process is exist now ">>$logFile
        exit 0
fi

#check start.sh
num=`ps -ef |grep '$bashPath/start.sh' |grep -v grep|wc -l`
echo "`date` the num of start.sh is $num">>$logFile
if [ $num -gt 1 ];then
        echo "the other start.sh is executing now ">>$logFile
        exit 0
fi

getConfCmd="curl http://sumeru.configplugin.wsd.com/getConfigContent -d env=$SUMERU_ENV&app=$SUMERU_APP&server=$SUMERU_SERVER&container_name=$SUMERU_CONTAINER_NAME -D $headerConfFile -o $trpcConfFile"

echo "the getConfCmd is  $getConfCmd">>$logFile

ret=`$getConfCmd`

if [ ! -f $headerConfFile ];then
    echo "cannot find header.conf after getConf">>$logFile
        rm $trpcConfFile
    exit $getConfFail
fi

if [ ! -f $trpcConfFile ];then
    echo "cannot find trpc.conf after getConf">>$logFile
    exit $getConfFail
fi

#judge the result
httpRet=`head -n 1 $headerConfFile |awk '{print $2}'`
if [ $httpRet -ne 200 ];then
   echo "call trpcConf http fail and it return $httpRet and the details config is `cat $trpcConfFile`">>$logFile
   rm $trpcConfFile $headerConfFile
   exit $getConfFail
fi

#judge interface code
interfaceCode=$(grep 'Code' $headerConfFile | awk '{print $2}'|awk -F '\r' '{print $1}')
if [ $interfaceCode -ne 0 ]; then
    errMsg=$(grep 'Msg' $headerConfFile | awk -F 'Msg' '{print $2}')
    echo "call trpcConf http success but assembly config fail ;the retCode is $interfaceCode and the errMsg is $errMsg">>$logFile
    exit $getConfFail
fi

confVersion=`cat $headerConfFile |grep -i version |awk -F ': ' '{print $2}'`
echo "the conf version is $confVersion">>$logFile
confMd5=`cat $headerConfFile |grep -i md5 |awk -F ': ' '{print $2}'| awk -F '\r' '{print $1}'`
echo "the conf md5 is $confMd5">>$logFile
fileMd5=`md5sum $trpcConfFile |awk '{print $1}'`
echo "the file md5 is $fileMd5">>$logFile

if [ "$confMd5" != "$fileMd5" ];then
    echo "the md5 of trpc conf is not equeal the result md5,so the file is broken">>$logFile
    rm $headerConfFile $trpcConfFile
    exit $getConfFail
fi

rm $headerConfFile

mv $trpcConfFile $trpcPath/trpc_go.yaml
echo "get trpc conf done" >>$logFile

#start process
bash -c "echo -n 'start at ' && date" >> /usr/local/app/start_history.log
ulimit -c unlimited

if [ ! -f $trpcPath/$SUMERU_SERVER ];then
     echo "the code file is not exist in $trpcPath" >>$logFile
     exit $fileMissing
fi
chmod a+x $trpcPath//$SUMERU_SERVER
cd $trpcPath
nohup $trpcPath/$SUMERU_SERVER -conf=$trpcPath/trpc_go.yaml  2>&1 | tee  $serverLogFile>>$serverHistoryLogFile 2>&1 &

sleep 3
num=`ps -ef |grep $trpcPath/$SUMERU_SERVER |grep -v grep|wc -l`
echo "the num of process after start is $num">>$logFile
if [ $num -lt 1 ];then
    echo "the process is not exit after start ">>$logFile
    exit $startProcessFail
fi
pid=`ps -ef |grep $trpcPath/$SUMERU_SERVER |grep -v grep|head -n 1 |awk '{print $2}'`
echo $pid>$pidConf
exit 0

