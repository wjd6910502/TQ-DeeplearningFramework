#!/bin/bash

set -x
ret=`service ssh start`
echo $ret

while :
do
    echo "I love you forever"
    sleep 2
done


#./horovod-service -s slave

#//for循环
