package slave

import (
  "encoding/json"
  "fmt"
  "strconv"
  "time"
  "os/exec"
  "os"
  "syscall"
  "bytes"
  "strings"
  config "server/service/tq-service/config"
  env "server/service/tq-service/env"
  httprequest "server/service/tq-service/httprequest"
  util "server/service/tq-service/util"
  resourcemgr "server/service/tq-service/resourcemanager"
  )

type StateType int32

const (
  State_IDLE StateType = 0
  State_RESOURCESCHEDULE_BEGIN StateType = 1
  State_RESOURCESCHEDULE_PROCESSING StateType = 2
  State_RESOURCESCHEDULE_FINISH StateType = 3
  State_TASK_START StateType = 4
  State_TASK_RUNNING StateType = 5
  State_TASK_FINISH StateType = 6
  State_MAX StateType = 7
  )

const MAX_TIMEOUT int64 = 300 //120s
const RETRY_COUNT int = 3 // 

type StateStrategy struct {
  taskId string
  ipCnt string
  slotCnt string
  cpuCnt string
  memSize string
  algType string
  batchsize string
  epochs string
  learningrate string
  param1 string
  param2 string
  retryCnt int
  applyTime int64
  bapplySuccess bool
  curState StateType  
  lastState StateType 
  savePath string     // model savepath
  trainPath string    // data loadpath
  hvdfilePath string  // hvd file path
  pid int             // pid
}

// 容错 状态回滚
func (s *StateStrategy) RollBackState(){
  s.curState  = s.curState - 1
  s.lastState = s.lastState - 1
}

// clear
func (s *StateStrategy) Clear(){
  s.taskId = ""
  s.ipCnt  = ""
  s.slotCnt = ""
  s.cpuCnt = ""
  s.memSize = ""
  s.algType = "lrd"
  s.batchsize = "32"
  s.epochs = "100"
  s.learningrate = "0.01"
  s.param1 = ""
  s.param2 = ""
  s.retryCnt = 0
  s.applyTime = 0
  s.bapplySuccess = false
  s.curState = State_IDLE
  s.lastState = State_IDLE
  s.savePath = ""
  s.trainPath = ""
  s.hvdfilePath = ""
  s.pid = 0
}

func (s *StateStrategy) UploadLog(){
  util.Info("Upload log................................")
  // 将训练完的日志上传到cos接口,通过任务id来查询训练过程.
  // cd /data1; mkdir taksid; ssh ip; copy ./log/debug-python.log /data1/taskid;
  shellPath := fmt.Sprintf("./upload_log.sh %v",s.taskId)
  cmd := exec.Command("/bin/bash","-c",shellPath)
  util.Infof("Exec Cmd Arg: %v , Process PID: %v ",cmd.Args,s.pid)

  // Run starts the specified command and waits for it to complete
  var out bytes.Buffer
  var stderr bytes.Buffer
  cmd.Stdout = &out
  cmd.Stderr = &stderr
  err := cmd.Run()
  if nil != err {
    util.Error(fmt.Sprintf("Cmd Go Wrong, err : %v , Process PID: %v",stderr.String(),cmd.Process.Pid))
    util.Error(fmt.Sprint(err) + ": " + stderr.String())
    return
  }

  //util.Infof("Cmd Arg: %v , Process PID: %v ",cmd.Args,s.pid)
  if out.String() == "" {
    util.Info("*****************Upload log Success********************* ")
  }else{
    util.Error("*****************Upload Unknow Wrong : " + out.String())
  }
}

func (s *StateStrategy) ReleaseResource(){
  util.Info("Release resource..........")

	headers := map[string]string{}
	params := map[string]string{}
	body := map[string]string{}

	headers["STAFFNAME"] = env.RESOURCE_APPLY_USER
	body["task_uuid"] = s.taskId

	util.Infof("reParams = %v", body)

	url := env.RESOURCE_APPLY_ADDR + env.RESOURCE_APPLY_PREFIX_RECYCLE_RESOURCE
	rep, err := httprequest.Post(url, body, params, headers)
	if err != nil{
		util.Infof("release resource err, taskId:%v, %v", s.taskId, err)
	}

	// !!!make sure rep1 is map
	_, rep1 := httprequest.Parse2Json(rep)
	if _, ok := rep1["error"]; ok{
		errcode := int(rep1["error"].(float64))
		if errcode == 0 {
			util.Infof("release resource success, taskId:%v, resp:%v", s.taskId, rep1)
		}else{
			util.Error(fmt.Sprintf("release resource task failed, taskId:%v, resp:%v, params:%v",s.taskId, rep1, body))
		}
	}else{
		util.Error(fmt.Sprintf("release resource exception, taskId:%v, resp:%v, params:%v",s.taskId, rep1, body))
	}
}

func (s *StateStrategy) RollBackStart(){
    // upload log
    s.UploadLog()
    // release applyresource
    s.ReleaseResource()
    //
    s.Clear()
}

func (s *StateStrategy) Init(){
    s.Clear()
}

func (s *StateStrategy) WriteState2Redis(){
  // update status for task
  util.Info("writestate2redis..............")
  headers := map[string]string{}
  body := map[string]string{}
  params := map[string]string{}

  body["operate"] = "set"
  body["field"] = "horovod-task"
  body["key"] = fmt.Sprintf("horovod:tq:task:status:%s", s.taskId)
  body["value"] = fmt.Sprint(s.curState)
  body["timeout"] = strconv.Itoa(24*60*60)

  url := env.REDIS_ADDR + env.REDIS_PREFIX
  headers["Authorization"] = env.COCO_TOKEN

  rep, _ := httprequest.Post(url, body, params, headers)
  _, rep1 := httprequest.Parse2Json(rep)
  util.Info(fmt.Sprintf("update taskStatus result, taskId:%v, resp:%v", s.taskId, rep1))
}

func (s *StateStrategy) UpdateState(){
  util.Info("UpdateState..............")
  s.lastState  = s.curState
  s.curState = (s.curState + 1)%State_MAX
  s.WriteState2Redis()
}

func (s *StateStrategy) GetTask2Coco(){
  util.Info("@@@@@@@@@@@@@ fetch data from coco....... ")
  s.Init()

  headers := map[string]string{}
    params := map[string]string{}
  url := env.COCO_ADDR + env.COCO_PREFIX_GET
  headers["Authorization"] = env.COCO_TOKEN

  rep, _ := httprequest.Get(url,params,headers)
  _,rep1 := httprequest.Parse2Json(rep)

  // !!! make sure rep1 is map
  util.Infof("fetch taskdata = %v ",rep1)
  if _, ok := rep1["error"]; !ok{
    return
  }
  retcode := int(rep1["error"].(float64))
  util.Infof("fetch taskdata retcode = %v",retcode)

  if retcode == 0{
    result := rep1["result"]

    var data map[string]string
    err := json.Unmarshal([]byte(result.(string)), &data)
    if err != nil {
      return
    }

    util.Infof("********data = %v ", data)
    s.taskId = data["Taskid"]
    s.ipCnt = data["Ipcnt"]
    s.slotCnt = data["Slotcnt"]
    s.cpuCnt = data["Cpucnt"]
    s.memSize  = data["Memsize"]
    s.algType = data["Algtype"]
    s.batchsize = data["Batchsize"]
    s.epochs = data["Epoch"]
    s.learningrate = data["Learningrate"]
    s.param1  = data["Param1"]
    s.param2  = data["Param2"]
    s.retryCnt = 0
    s.applyTime = 0
    s.bapplySuccess = false
    s.savePath = data["Savepath"]
    s.trainPath = data["Loadpath"]
    s.hvdfilePath = "./ip.json"
    util.Infof("********e = %v", s)

    s.UpdateState()
  }
  //else do nothing
}

func (s *StateStrategy) ApplyResource(cfgparams config.ResouceApplyParams) bool{
  util.Info(fmt.Sprintf("start apply resource, taskId:%s", s.taskId))

  if s.retryCnt > RETRY_COUNT{
    s.RollBackStart()
    return false
  }

  headers := map[string]string{}
  params := map[string]string{}

  headers["STAFFNAME"] = env.RESOURCE_APPLY_USER
  body, err := GenerateApplyParams(s, cfgparams)
  if err != nil{
    s.retryCnt = s.retryCnt + 1
    util.Error(fmt.Sprintf("resourceApplyParams error:%v", err))
    return false
  }
  util.Infof("applyParams = %v", body)

  url := env.RESOURCE_APPLY_ADDR + env.RESOURCE_APPLY_PREFIX_CREATE
  rep, err := httprequest.Post(url, body, params, headers)
  if err != nil{
    util.Infof("applyResource err, taskId:%v, %v", s.taskId, s.retryCnt)
    s.retryCnt = s.retryCnt + 1
    return false
  }

  // !!!make sure rep1 is map
  _, rep1 := httprequest.Parse2Json(rep)
  if _, ok := rep1["error"]; ok{
    errcode := int(rep1["error"].(float64))
    if errcode == 0 {
      util.Infof("create task success, taskId:%v, resp:%v", s.taskId, rep1)
    }else{
      util.Infof("create task failed, taskId:%v, resp:%v, applyParams:%v",s.taskId, rep1, body)
      s.retryCnt = s.retryCnt + 1
      return false
    }
  }else{
    util.Infof("create task exception, taskId:%v, resp:%v, applyParams:%v",s.taskId, rep1, body)
    s.retryCnt = s.retryCnt + 1
    return false
  }
  s.UpdateState()

  // set start-apply time
  s.applyTime = time.Now().Unix()
  return true
}

func (s *StateStrategy) CheckApplyStatus() bool{
  util.Infof("CheckApplyStatus has spend %v s",(time.Now().Unix() - s.applyTime))
  if (time.Now().Unix() - s.applyTime) > MAX_TIMEOUT{
      s.RollBackStart()
      return false
  }

  headers := map[string]string{}
  params := map[string]string{}
  body := map[string]string{}
  body["app_id"] = s.taskId
  headers["STAFFNAME"] = env.RESOURCE_APPLY_USER
  url := env.RESOURCE_APPLY_ADDR + env.RESOURCE_APPLY_PREFIX_QUERY_STATE
  rep, err := httprequest.Post(url, body, params, headers)

  if err != nil{
    util.Error(fmt.Sprintf("check err, taskId:%s, %d ", s.taskId, ))
    return false
  }
  _, rep1 := httprequest.Parse2Json(rep)
  if _, ok := rep1["error"]; !ok{
    util.Error(fmt.Sprintf("get task_status exception, taskId:%v, resp:%v",s.taskId, rep1))
    return false
  }

  errcode := int(rep1["error"].(float64))
  if errcode != 0 {
    util.Error(fmt.Sprintf("get task_status failed, taskId:%v, resp:%v",s.taskId, rep1))
    return false
  }
  data := rep1["data"].(map[string]interface{})
  if _, ok := data["task_state"]; !ok{
    util.Error(fmt.Sprintf("no task_state in result, taskId:%v, resp:%v", s.taskId, rep1))
    return false
  }
  status := int(data["task_state"].(float64))
  if status != 1 {
    util.Error(fmt.Sprintf("task status is not ready:%v, resp:%v", s.taskId, data))
    return false
  }
  ips := data["address_list"].([]interface{})
  runningNum := int(data["running_num"].(float64))
  ipCnt, err := strconv.Atoi(s.ipCnt)
  if err != nil || runningNum != ipCnt{
    util.Error(fmt.Sprintf("runningNum:%d is not equal ipCnt:%d or err:%v,data:%v", runningNum, ipCnt, err, data))
    return false
  }
  util.Infof("get ips success in check status, ips:%v, hvdfilePath:%s,slotCnt:%s", ips, s.hvdfilePath, s.slotCnt)
  err = SaveIpsToFile(ips, s.slotCnt, s.hvdfilePath)
  if err != nil{
    util.Error(fmt.Sprintf("save ips err:%v in check status, ips:%v, hvdfilePath:%s,slotCnt:%s", err, ips, s.hvdfilePath, s.slotCnt))
    return false
  }
  s.UpdateState()
  s.bapplySuccess = true
  return true
}

// split_data
func (s *StateStrategy) PreprocesingData(){
  util.Infof("start preprocesingData and split")
  ipCnt, err_code := strconv.Atoi(s.ipCnt)
  slotCnt, err_code1 := strconv.Atoi(s.slotCnt)
  if nil != err_code || nil != err_code1 {
      util.Error(fmt.Sprintf("Start Params Wrong, ipCnt %s, slotCnt %s", s.ipCnt,s.slotCnt))
      s.RollBackStart()
      return
  }

  sCnt := ipCnt*slotCnt
  loadertype,err := strconv.Atoi(s.param1)
  if nil != err {
      util.Error(fmt.Sprintf("Start Params Wrong, ipCnt %s, slotCnt %s", s.ipCnt,s.slotCnt))
      s.RollBackStart()
      return
  }

  trainpath := resourcemgr.PrepareData(sCnt,s.trainPath,loadertype)
  if trainpath == "" {
    s.RollBackStart()
    return
  }
  s.trainPath = trainpath
}

/* start horovod shell, params:
       1. np
       2. algorthm
       3. epochs
       4. batchsize
       5. learningrate
       6. trainPath
       7. savePath
       8. taskid
       9. report_ip 
      10. report_token
*/
func (s *StateStrategy) StartScript(){
  util.Info("*****************Begin Start up Script ********************* ")
  if s.bapplySuccess == false {
    s.RollBackStart()
    return
  }

  ipCnt, err_code := strconv.Atoi(s.ipCnt)
  slotCnt, err_code1 := strconv.Atoi(s.slotCnt)
  if nil != err_code || nil != err_code1 {
      util.Error(fmt.Sprintf("Start Params Wrong, ipCnt %s, slotCnt %s", s.ipCnt,s.slotCnt))
  }

  // compute np
  np := ipCnt*slotCnt
  shellPath := fmt.Sprintf("./test.sh %d %s %s %s %s %s %s %v %v %v",np,s.algType,s.epochs,s.batchsize,s.learningrate,s.trainPath,s.savePath,s.taskId,env.REPORT_ADDR,env.REPORT_TOKEN)
  util.Infof("shellpath = %v",shellPath)
  cmd := exec.Command("/bin/bash","-c",shellPath) //Cmd init
  util.Infof("Start Exec Cmd Arg: %v , Process PID: %v ",cmd.Args,s.pid)

  // get local env
  cmd.Env = s.GetCmdEnv()

  // Run starts the specified command and waits for it to complete
  var out bytes.Buffer
  var stderr bytes.Buffer
  cmd.Stdout = &out
  cmd.Stderr = &stderr
  err := cmd.Run()
  if nil != err {
    util.Error(fmt.Sprintf("Cmd Go Wrong, err : %v",stderr.String()))
    util.Error(fmt.Sprint(err) + ": " + stderr.String())
    s.RollBackStart()
    return
  }

  s.pid = cmd.Process.Pid
  util.Infof("Cmd Arg: %v , Process PID: %v ",cmd.Args,s.pid)

  if out.String() == "" {
    util.Info("*****************Start up Script Success********************* ")
  }else{
    util.Error("*****************Start up Script Unknow Wrong : " + out.String())
    s.RollBackStart()
    return
  }

  s.UpdateState()
}

func (s *StateStrategy) GetCmdEnv() []string {
  env := os.Environ()
  cmdEnv := []string{}

  for _, e := range env {
    i := strings.Index(e, "=")
    if i > 0 && (e[:i] == "ENV_NAME") {
       // do yourself
    } else {
       cmdEnv = append(cmdEnv, e)
    }
  }

  return cmdEnv
}

func (s *StateStrategy) CheckPid() bool {
  util.Info("checkpid success......... ")
  process, err := os.FindProcess(s.pid)
  if err != nil {
    util.Info(fmt.Sprintf("Unable to find the process %d", s.pid))
    return false
  }

  err = process.Signal(syscall.Signal(0))
  //util.Info(err)
  if err != nil {
    util.Infof("Process %d is dead!", s.pid)
    return false
  } else {
    util.Infof("Process %d is alive!", s.pid)
    return true
  }
}

func (s *StateStrategy) CheckRuntimeState(){
  util.Infof("CheckRuntimeState........pid = ",s.pid)
  if s.CheckPid() == true {
    util.Info("time spend ")
  }else{
    util.Infof("CheckApplyStatus has spend %vs",(time.Now().Unix() - s.applyTime))
    s.UpdateState()
  }
}

func (s *StateStrategy) FinishTaskState(){
  util.Info("FinishTaskState........ ")
  s.UpdateState()
  // upload log
  s.UploadLog()
  // release applyresource
  s.ReleaseResource()
}

func (s *StateStrategy)  GetState() StateType {
  return s.curState
}

func FSM() int32 {
  s :=  StateStrategy{curState:0}
  util.Initialize()
  // init resource apply default params
  cfgparams := config.ResouceApplyParams{}
  cfgparams.LoadTflowCfg(env.RESOURCE_APPLY_CONFIG_PATH)

  for true {
    switch s.GetState() {
      case State_IDLE:
        util.Info("state 0")
        s.GetTask2Coco()
        time.Sleep(time.Duration(2)*time.Second)
        break
      case State_RESOURCESCHEDULE_BEGIN:
        util.Info("state 1")
        s.ApplyResource(cfgparams)
        time.Sleep(time.Duration(2)*time.Second)
        break
      case State_RESOURCESCHEDULE_PROCESSING:
        util.Info("state 2")
        s.CheckApplyStatus()
        time.Sleep(time.Duration(2)*time.Second)
        break
      case State_RESOURCESCHEDULE_FINISH:
        util.Info("state 3")
        s.PreprocesingData()//处理数据
        s.UpdateState()
        time.Sleep(time.Duration(2)*time.Second)
        break
      case State_TASK_START:
        util.Info("state 4")
        s.StartScript()
        time.Sleep(time.Duration(2)*time.Second)
        break
      case State_TASK_RUNNING:
        util.Info("state 5")
        s.CheckRuntimeState()
        time.Sleep(time.Duration(2)*time.Second)
        break
      case State_TASK_FINISH:
        util.Info("state 6")
        s.FinishTaskState()
        time.Sleep(time.Duration(2)*time.Second)
        break
      default:
        time.Sleep(time.Duration(2)*time.Second)
      }
  }
  return 0
}
