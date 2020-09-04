package slave

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net"
  "os"
  "path/filepath"
  "server/service/tq-service/config"
  "server/service/tq-service/env"
  util "server/service/tq-service/util"
)

func GenerateApplyParams(s *StateStrategy, applyPamas config.ResouceApplyParams) (map[string]string, error) {
  jsonBytes, err := json.Marshal(applyPamas)
  if err != nil{
    return nil, err
  }
  paramsMap := map[string]string{}
  if err := json.Unmarshal([]byte(jsonBytes), &paramsMap); err != nil {
    return paramsMap, err
  }

  paramsMap["app_id"] = s.taskId
  if s.ipCnt != ""{
    paramsMap["replicas"] = s.ipCnt
  }
  if s.cpuCnt != "" {
    paramsMap["cpu_core"] = s.cpuCnt
  }
  if s.memSize != "" {
    paramsMap["cpu_memory"] = fmt.Sprintf("%sGi", s.memSize)
  }
  envVariablesMap := map[string]string{
    "TRAIN_DATA_COLLECTION_SERVER": env.COCO_ADDR + env.COCO_PREFIX_TRAIN_DATA_COLLECTION,
    "COCO_TOKEN": env.COCO_TOKEN,
    "TASK_ID": s.taskId,
    "REPORT_ADDR":env.REPORT_ADDR,
    "REPORT_TOKEN":env.REPORT_TOKEN,
    "S3_ENDPOINT":env.S3_ENDPOINT,
    "S3_BUCKET":env.S3_BUCKET,
    "S3_ACCESSKEY":env.S3_ACCESSKEY,
    "S3_SECRETKEY":env.S3_SECRETKEY}

  jsonByte, err := json.Marshal(envVariablesMap)
  if err != nil{
    paramsMap["extend_env_variables"] = "{}"
    util.Error(fmt.Sprintf("envVaiableMap convert to str err %v", err))
    return paramsMap, err
  }else{
    paramsMap["extend_env_variables"] = string(jsonByte)
  }
  return paramsMap, nil
}

/* save ips file format：
    aa slots=2
    bb slots=2
    cc slots=2
*/
func SaveIpsToFile(ips []interface{}, slots string, filePath string) (error) {

  dir, err := filepath.Abs(filepath.Dir(filePath))
  isOk, err := PathExists(dir)
  if !isOk{
    err = os.MkdirAll(dir, 0777)
  }

  var tmp_str = ""
  for _, ip := range ips{
    ipStr := ip.(string)
    address := net.ParseIP(ipStr)
    if address == nil {
      util.Error(fmt.Sprintf("ip parse error: %v, cannot run horovod", ip))
    }
    tmp_str = fmt.Sprintf("%s slots=%s\n%s", ipStr, slots, tmp_str)
  }
  err = ioutil.WriteFile(filePath, []byte(tmp_str), 0644)
  if err != nil{
    return err
  }
  return nil
}

func PathExists(path string) (bool, error) {
  _, err := os.Stat(path)
  if err == nil {
    return true, nil
  }
  if os.IsNotExist(err) {
    return false, nil
  }
  return false, err
}
