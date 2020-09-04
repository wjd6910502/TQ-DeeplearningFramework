
import os

#S3_ENDPOINT  = os.environ.get("S3_ENDPOINT","http://lingqu.cephrados.so.db:7480")
#S3_ACCESSKEY = os.environ.get("S3_ACCESSKEY","adnywang-turinglab-platform-7e87487e")
#S3_SECRETKEY = os.environ.get("S#_SECRETKEY","adnywang-turinglab-platform-8debaa21")
S3_ENDPOINT  = os.environ.get("S3_ENDPOINT")
S3_ACCESSKEY = os.environ.get("S3_ACCESSKEY")
S3_SECRETKEY = os.environ.get("S3_SECRETKEY")

REPORT_ADDR   = os.environ.get("REPORT_ADDR","http://127.0.0.1") 
REPORT_TOKEN  = os.environ.get("REPORT_TOKEN","12345678987654321") 
REPORT_TASKID = os.environ.get("REPORT_TASKID","1")
