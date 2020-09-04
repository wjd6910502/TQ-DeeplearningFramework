#!/bin/bash

set -x
cp -rf ip.json myhostfile >> test.log

echo "np ="$1  >>test.log
echo "algtype = "$2 >>test.log
echo "epochs = "$3 >> test.log
echo "batchsize = "$4 >> test.log
echo "lrate = "$5 >> test.log
echo "load_path = "$6 >> test.log
echo "save_path = "$7 >> test.log



i=0
max=10
while :
do
	echo "one loop ..." >> test.log
	
	if [ $i -gt $max ]
	then
		exit
	fi
	i=$[$i+1];
	sleep 1
done

