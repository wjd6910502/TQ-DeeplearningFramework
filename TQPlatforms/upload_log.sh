#!/bin/bash

set -x
#echo "************** begin upload log **************" >> ./log/shell.log

taskid=$1
#echo "taskid = "$taskid
# parse myhostfile

#resdir=/app/dmlp-platform/log/*.log
#destdir=/data1/$taskid

resdir=./log/debug*.log
destdir=$taskid
if [ ! -x "$destdir" ]; then
  mkdir "$destdir"
fi

while read line
do
    ip=`echo $line | awk '{print $1}'`
    #echo "ip = "$ip
    #result=`ssh $ip "cp -rf $reddir $destdir"`
    cp -rf $resdir $destdir
    #echo $result
    #cat $destdir/*.log
done < myhostfile

#echo "************** end upload log **************" >> ./log/shell.log
