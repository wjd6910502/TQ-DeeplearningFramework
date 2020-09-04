package config

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "os"
)

type ResouceApplyParams struct {
  ImageName string `json:"image_name"`     				// 必选，指定训练模型的镜像
  ImageVersion string `json:"image_version"`  			// 可选，默认为latest， image_version指定镜像的版本，若为lastest, 则使用m-v1.1.1.1格式的最新版本
  StartCmd string `json:"start_cmd"`   					// 必选， start_cmd, 启动docker脚本, sleep infinity
  Label string `json:"label"`  							// 可选，指定tenc的label，默认为omt
  ConfigId string `json:"config_id"` 						// 可选，cephfs集群id，测试服默认为华南集群，正式服默认为上海V2集群
  SetId string `json:"set_id"`  							// 可选，服务集群id，测试服默认为华南集群，正式服默认为上海V2集群
  Namespace string `json:"namespace"` 					// 可选，默认测试服为turinglabtest，默认正式服为turiglab，可额外指定
  Replicas string `json:"replicas"`						// 必选， 分布式训练该项必选，否则只会拉起一个副本。
  CpuCore string `json:"cpu_core"`						// 可选，默认为4核
  CpuMemory string `json:"cpu_memory"`							// 可选， 容器内存限额，如100Mi, 1Gi等，默认为8Gi
  GpuNum string `json:"gpu_num"`							// 可选， 当为0时，表示只用cpu，> 0时分配GPU
  ExtendEnvVariables string `json:"extend_env_variables"` // 可选，该字段应该可以解析为dict类型，强烈建议填写，至少填写日志路径，便于在kibana上查看 demo: ':{"TENC_FILELOG_PATHS": “log_path”}' 扩展环境变量，可指定训练的数据存储路径，模型的包括路径，
  CephPath string `json:"ceph_path"`   					// 必选，主要是为了与cos的数据进行ceph交互。
  MountPath string `json:"mount_path"`  					// 必选，挂载到容器的路径，建议挂载到/data1下，便于与cos进行交互
  ProjectName string `json:"project_name"`   				// 必选， 表示项目名
}

func (rap *ResouceApplyParams) LoadTflowCfg(filePath string){
  //jsonFile, err := ReadJsonFile(filePath)
  jsonFile, err := os.Open(filePath)
  defer jsonFile.Close()
  // if we os.Open returns an error then handle it
  if err != nil {
    errStr := fmt.Sprintf("openfile:%s error:%v", filePath, err)
    log.Fatal(errStr)
  }
  log.Println(fmt.Sprintf("Successfully Opened %s file", filePath))
  // read opened jsonFile as a byte array.
  byteValue, _ := ioutil.ReadAll(jsonFile)
  json.Unmarshal(byteValue, rap)
}
