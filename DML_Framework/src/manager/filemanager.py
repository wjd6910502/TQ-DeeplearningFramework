import os
import sys
import boto3
from botocore.client import Config
from util.logger import logger

import tarfile

class Filemanager:
    def __init__(self):
        pass

    @staticmethod 
    def tar(src_dir,target_file):
        #创建压缩包名
        tar = tarfile.open(target_file,"w:gz")
        #创建压缩包
        for root,dir,files in os.walk(src_dir):
            for file in files:
                fullpath = os.path.join(root,file)
                tar.add(fullpath)
        tar.close()     
    
    @staticmethod
    def untar(tar_path, target_path):
        logger.info("untar 1111111111111111111111")
        try:
            tar = tarfile.open(tar_path, "r:gz")
            file_names = tar.getnames()
            for file_name in file_names:
                tar.extract(file_name, target_path)
            logger.info("untar filename =")
            tar.close()
        except:
            prin("err =",e) 

    @staticmethod
    def createfile(path):
        logger.info("1111111111111111111111")
        isExists=os.path.exists(path)
        if not isExists: 
            os.makedirs(path) 
            logger.info("create dir success")

        return True
'''
#Filemanager.untar("../../1.tar.gz","./")
'''
