
package env

import (
  //"log"
  "os"
  util "server/service/tq-service/util"
)

var COCO_ADDR string
var COCO_PREFIX_CREATE string
var COCO_PREFIX_GET string
var COCO_TOKEN string
var COCO_PREFIX_TRAIN_DATA_COLLECTION string

var REDIS_ADDR string
var REDIS_PREFIX string

var RESOURCE_APPLY_ADDR string
var RESOURCE_APPLY_CONFIG_PATH string
var RESOURCE_APPLY_PREFIX_CREATE string
var RESOURCE_APPLY_PREFIX_QUERY_STATE string
var RESOURCE_APPLY_PREFIX_RECYCLE_RESOURCE string
var RESOURCE_APPLY_USER string
var RESOURCE_APPLY_PROJECT_NAME string      // 固定的就可以，这样方便后续分配资源时采取默认值

var REPORT_ADDR string
var REPORT_TOKEN string

var TOSS_ADDR string
var TOSS_PREFIX string

var S3_ENDPOINT string
var S3_BUCKET string
var S3_ACCESSKEY string
var S3_SECRETKEY string

func Init() {
  LoadEnv()
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        os.Setenv(key,defaultValue)
        return defaultValue
    }
    return value
}

func LoadEnv(){

  util.Info("start env......")

  COCO_ADDR = getEnv("COCO_ADDR","")
  COCO_PREFIX_GET = getEnv("COCO_PREFIX_GET","/coco/task/horovod/dmlp-schedule/1/get")
  COCO_PREFIX_CREATE = getEnv("COCO_PREFIX_CREATE","/coco/task/horovod/dmlp-schedule/create")
  COCO_TOKEN = getEnv("COCO_TOKEN","Token ")
  COCO_PREFIX_TRAIN_DATA_COLLECTION = getEnv("COCO_PREFIX_TRAIN_DATA_COLLECTION","/coco/task/horovod-train/collection/create")

  RESOURCE_APPLY_ADDR = getEnv("RESOURCE_APPLY_ADDR"," ")
  RESOURCE_APPLY_CONFIG_PATH = getEnv("RESOURCE_APPLY_CONFIG_PATH","/app/dmlp-platform/config/tflow-dev.json")
  RESOURCE_APPLY_PREFIX_CREATE = getEnv("RESOURCE_APPLY_PREFIX_CREATE","/res/create")
  RESOURCE_APPLY_PREFIX_QUERY_STATE = getEnv("RESOURCE_APPLY_PREFIX_QUERY_STATE","/res/query_service_pod_info")
  RESOURCE_APPLY_PREFIX_RECYCLE_RESOURCE = getEnv("RESOURCE_APPLY_PREFIX_RECYCLE_RESOURCE","/res/recycle_resource")
  RESOURCE_APPLY_USER = getEnv("RESOURCE_APPLY_USER","tq")
  RESOURCE_APPLY_PROJECT_NAME = getEnv("RESOURCE_APPLY_PROJECT_NAME","tq")

  REDIS_ADDR = getEnv("REDIS_ADDR"," ")
  REDIS_PREFIX = getEnv("REDIS_PREFIX","/redis/task/redis/op")

  REPORT_ADDR = getEnv("REPORT_ADDR"," ")
  REPORT_TOKEN = getEnv("REPORT_TOKEN"," ")

  TOSS_ADDR = getEnv("TOSS_ADDR","")
  TOSS_PREFIX = getEnv("TOSS_PREFIX"," ")

  S3_ENDPOINT = getEnv("S3_ENDPOINT"," ")
  S3_BUCKET = getEnv("S3_BUCKET"," ")
  S3_ACCESSKEY = getEnv("S3_ACCESSKEY"," ")
  S3_SECRETKEY = getEnv("S3_SECRETKEY"," ")

}



