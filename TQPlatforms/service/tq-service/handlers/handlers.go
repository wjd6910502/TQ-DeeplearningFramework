package handlers

import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"reflect"
	pb "server/proto"
	env "server/service/tq-service/env"
	httprequest "server/service/tq-service/httprequest"
	util "server/service/tq-service/util"
)

// NewService returns a naïve, stateless implementation of Service.
func NewService() pb.TqServer {
	return tqService{}
}

type tqService struct{}

// CreateTask implements Service.
func (s tqService) CreateTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	var resp pb.TaskResponse
  // TODO:params check
  // savepath loadpath check
  // 其他参数验证
  if in.Algtype == "" || in.Loadpath == "" || in.Savepath == "" || in.Slotcnt == "" || in.Ipcnt == "" || in.Taskid == "" {
     log.Println("Lack of Params")
     resp = pb.TaskResponse{
		  in.Taskid, // TaskId:
			"-1",      // Retcode:/log
			"0", // ErrCode:
		 }
     return &resp,nil
  }

  if in.Memsize == ""{
    in.Memsize = "8"
  }
  if in.Batchsize == ""{
    in.Batchsize = "32"
  }
  if in.Epoch == ""{
    in.Epoch ="1"
  }
  if in.Learningrate == ""{
    in.Learningrate = "0.1"
  }

	//将任务插入到coco队列
	headers := map[string]string{}
	body := map[string]string{}
	params := map[string]string{}
	message := map[string]string{}

	url := env.COCO_ADDR + env.COCO_PREFIX_CREATE
	headers["Authorization"] = env.COCO_TOKEN

	//optiomize TODO: load json
	elem := reflect.ValueOf(in).Elem()
	relType := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		message[relType.Field(i).Name] = elem.Field(i).Interface().(string)
	}

	message["TaskId"] = in.Taskid
	message["Ipcnt"] = in.Ipcnt
	util.Infof("message =", message)
	json_str, err := json.Marshal(message)
	if err != nil {
		util.Infof("json.Marshal failed:", err)
		resp = pb.TaskResponse{
			in.Taskid, // TaskId:
			"-1",      // Retcode:/log

			"0", // ErrCode:
		}
		return &resp, nil
	}

	util.Infof("json_str =", string(json_str))
	body["message"] = string(json_str)
	rep, _ := httprequest.Post(url, body, params, headers)
	util.Infof("rep =", rep)
	_, rep1 := httprequest.Parse2Json(rep)

	util.Infof("@@@@@@@@@@@ CREATETASK", rep1)
	retcode := httprequest.Float642String(rep1["error"].(float64))

	resp.TaskId = in.Taskid
	resp = pb.TaskResponse{
		in.Taskid, // TaskId:
		retcode,   // Retcode:
		"0",       // ErrCode:
	}
	return &resp, nil
}

// GetTaskStatus implements Service.
func (s tqService) GetTaskStatus(ctx context.Context, in *pb.TaskStatusRequest) (*pb.TaskStatusResponse, error) {
	var resp pb.TaskStatusResponse
	util.Infof("@@@@@@@@@@ gettaskstatus")
	// 从redis读取状态 #TODO
	//httprequest.get
	headers := map[string]string{}
	body := map[string]string{}
	params := map[string]string{}

	body["operate"] = "get"
	body["field"] = "horovod-task"
	body["key"] = fmt.Sprintf("horovod:tq:task:status:%s", in.TaskId)

	url := env.REDIS_ADDR + env.REDIS_PREFIX
	headers["Authorization"] = env.COCO_TOKEN

	rep, _ := httprequest.Post(url, body, params, headers)
	_, rep1 := httprequest.Parse2Json(rep)

	util.Info(fmt.Sprintf("get taskStatus result, taskId:%s, resp:%s", in.TaskId, rep1))

	retcode := httprequest.Int642String(int64(rep1["Errcode"].(float64)))
	status := "-1"
	if retcode != "-1" {
		status = rep1["Msg"].(string)
	}
	resp = pb.TaskStatusResponse{
		in.TaskId, // TaskId:
		status,    // Status:
		retcode,   // ErrCode:
	}
	return &resp, nil
}
