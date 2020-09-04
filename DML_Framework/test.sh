#!/bin/bash

set -x
cp -rf ip.json myhostfile >> test.log

np=$1
algtype=$2
epochs=$3
batchsize=$4
lrate=$5
loadpath=$6
savepath=$7
taskid=$8
r_addr=$9
r_token=${10}


rm -rf ./log/shell.log
touch ./log/shell.log

echo "np = "$np >>./log/shell.log
echo "algtype = "$algtype >>./log/shell.log
echo "epochs = "$epochs >>./log/shell.log
echo "batchsize = "$batchsize >>./log/shell.log
echo "lrate = "$lrate >>./log/shell.log
echo "loadpath = "$loadpath >>./log/shell.log
echo "savepath = "$savepath >>./log/shell.log
echo "taskid = "$taskid >>./log/shell.log
echo "addr = "$r_addr >>./log/shell.log
echo "token = "$r_token >>./log/shell.log

# TODO 同步python脚本和验证环境变量

# optimize 
# --fusion-threshold-mb=64M  
# --timeline-filename /path/to/timeline.json --timeline-mark-cycles
# --autotune --autotune-log-file /tmp/autotune_log.csv

horovodrun -np $np -hostfile myhostfile -p 10234 python3 train.py -alg_type $algtype -epochs $epochs -batch_size $batchsize -lrate $lrate -train_path $loadpath -save_path $savepath -taskid $taskid -report_addr $r_addr -report_token $r_token  >> ./log/debug-python.log
