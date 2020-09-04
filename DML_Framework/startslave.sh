#!/bin/bash

set -x

ret=`service ssh start`
echo $ret

./tqservice -s slave >> test.log




