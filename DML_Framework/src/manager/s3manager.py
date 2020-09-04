
import os
import sys
import boto3
from botocore.client import Config
import env.env as env 
from util.logger import logger

class S3mgr:

    def __init__(self):
        
        logger.info("url = {},s3acesskey = {} ,s3secretkey = {}".format(env.S3_ENDPOINT,env.S3_ACCESSKEY,env.S3_SECRETKEY))
        self.access_key = env.S3_ACCESSKEY
        self.secret_key = env.S3_SECRETKEY
        self.endpoint = env.S3_ENDPOINT
        logger.info("url = {},s3acesskey = {} ,s3secretkey = {}".format(self.endpoint,self.access_key,self.secret_key)) 
        self.client = boto3.client('s3', endpoint_url=self.endpoint,
                         aws_access_key_id=self.access_key,
                         aws_secret_access_key=self.secret_key,
                         region_name='cn-north-1',
                         config=Config(signature_version='s3'))
    
    def list_buckets(self):
        bucket_list = self.client.list_buckets()
        return bucket_list

    def create_buckets(self,name):
        bucket = self.client.create_bucket(Bucket=name)
        return bucket
    
    def get_bucket(self,name):
        bucket = self.client.head_bucket(Bucket=name)
        return bucket
    
    def del_bucket(self,name):
        bucket = self.client.delete_bucket(Bucket=name)
        return bucket
    
    def get_bucket_list(self,name):
        bucket = self.client.list_objects(Bucket=name)
        return bucket
    
    def upload_file(self,bucket,srcname,destname):
        self.client.upload_file(srcname,bucket,destname,ExtraArgs={'ACL':'public-read'})

    def download_file(self,bucket,srcfile,destfile):
        if bucket == "":
            self.client.download_file("s3__turinglab-platform",srcfile,destfile)
        else:
            self.client.download_file(bucket,srcfile,destfile)
'''        
s3 = S3mgr()
blist = s3.list_buckets()
print("blist = ",blist)

blist = s3.get_bucket_list("s3__turinglab-platform")
print("blist = ",blist)

s3.download_file("s3__turinglab-platform","slice_57640ece-c678-11ea-a45d-7af6424c9416_1595841231_0.tar.gz","/app/dmlp-platform/tmpdata.tar.gz")
'''
