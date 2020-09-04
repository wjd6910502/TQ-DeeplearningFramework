#!/bin/bash

set -x
#echo "************** begin upload log **************" >> ./log/shell.log

taskid=$1

resdir=/app/dmlp-platform/log/debug-python.log
destdir=/data1/tq/log/$taskid
#resdir=./log/debug*.log
#destdir=$taskid

if [ ! -x "$destdir" ]; then
  mkdir "$destdir"
fi

while read line
do
    ip=`echo $line | awk '{print $1}'`
    result=`ssh $ip "mv $resdir $destdir/$ip.log"`
    #$cp -rf $resdir $destdir/$ip.log
done < myhostfile

#echo "************** end upload log **************" >> ./log/shell.log
