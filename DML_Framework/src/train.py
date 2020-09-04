
import os
import sys
import string

from manager.algorthmgr import Algothmgr 
from util.logger import logger
import argparse

import threading

# data parse
def parse_args():
    parser = argparse.ArgumentParser(description='Horovod Runner')
    parser.add_argument('-train_path', '--train_path', action='store', dest='train_path', help='training path.')
    parser.add_argument('-save_path', '--save_path', action='store', dest='save_path', help='save path.')
    parser.add_argument('-alg_type', '--alg_type', action='store', dest='alg_type', required=True,help='alg type.')
    parser.add_argument('-param_initail', '--param_initial',action='store', dest='param_initial', help='param_initial')
    parser.add_argument('-op', '--op',action='store', dest='op', help='op: adgsum,sdg..')
    parser.add_argument('-epochs', '--epochs',action='store', dest='epochs', help='epoch setting')
    parser.add_argument('-batch_size', '--batch_size',action='store', dest='batch_size', help='batchsize setting')
    parser.add_argument('-lrate','--lrate',action='store', dest='lrate', help='lrate setting')
    
    parser.add_argument('-taskid', '--taskid',action='store', dest='taskid', help='taskid setting')
    parser.add_argument('-report_addr','--report_addr',action='store', dest='report_addr', help='lrate setting')
    parser.add_argument('-report_token', '--report_token',action='store', dest='report_token', help='batchsize setting')


    args = parser.parse_args()
    return args

# data load    
def main():

    # 1
    args = parse_args()    
    logger.info("type = %s",args.alg_type) 
    
    #2
    mgr = Algothmgr(args)
    model = mgr.create_algorithm(args.alg_type, args)
    model.run()

import time
if __name__=="__main__":
    main()

