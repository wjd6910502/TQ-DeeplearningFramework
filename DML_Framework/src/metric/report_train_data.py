import requests
import json
import time
from util.logger import logger
from env import env

'''
  param msg_type: int(1-4), describe:
     MSG_TYPE_TRAIN_INFO = 1, kwargs need contain: epoch_num, batch_size, lr etc params
     MSG_TYPE_EPOCH_INFO = 2, kwargs need contain: current_epoch, time_used_per_epoch, loss, optional: acc, valid_loss, valid_acc
     MSG_TYPE_BATCH_INFO = 3, kwargs need contain: current_batch, time_used_per_batch, batch_steps
     MSG_TYPE_TRAIN_END = 4, kwargs need contain: train_end
'''
MSG_TYPE_TRAIN_INFO = 1
MSG_TYPE_EPOCH_INFO = 2
MSG_TYPE_BATCH_INFO = 3
MSG_TYPE_TRAIN_END = 4

def check_env(env_name, env_value, check_value):
    logger.error("no set {} in env".format(env_name)) if env_value == check_value \
        else logger.debug("env {}: {}".format(env_name, env_value))

def get_report_info():
    COCO_HEADERS = {"Authorization": "%s" % env.REPORT_TOKEN if env.REPORT_TOKEN.startswith("Token")
    else "Token %s" % env.REPORT_TOKEN}
    return env.REPORT_TASKID, env.REPORT_ADDR, COCO_HEADERS

'''
content:
    args: need contain epoch_num or current_epoch or current_batch or train_end
    result: bool, err_str
'''
def train_data_report(**kwargs):
    
    #TODO check_env
    is_ok, err_str = True, ""
    if not kwargs.__contains__("epochs") and not kwargs.__contains__("current_epoch") and \
            not kwargs.__contains__("current_batch") and not kwargs.__contains__("train_end"):
        raise ValueError('kwargs need contain ' +
                         'epoch_num(for train info report) or ' +
                         'current_epoch(for epoch info report) or' +
                         'current_batch(for batch info report) or' +
                         'train_end for report train ended')

    if kwargs.__contains__("train_end"):
        msg_type = MSG_TYPE_TRAIN_END
    elif kwargs.__contains__("current_batch"):
        msg_type = MSG_TYPE_BATCH_INFO
    elif kwargs.__contains__("current_epoch"):
        msg_type = MSG_TYPE_EPOCH_INFO
    else:
        msg_type = MSG_TYPE_TRAIN_INFO

    task_id, report_server, headers = get_report_info()
    data_json = {"msg_type": msg_type, "task_id": task_id}
    for k, v in kwargs.items():
        data_json[k] = str(v)
        
    st = time.time()

    try:
        resp = requests.post(report_server, headers=headers, timeout=3,
                             json={"message": json.dumps(data_json)})
    except Exception as err:
        err_str = "report train_data to remote_monitor exception, err:{}, " \
                  "url:{}".format(err, report_server)
        logger.error(err_str)
        is_ok = False
    else:
        result = json.loads(resp.text)
        if resp.status_code != 200 or result.get("error") == -1:
            err_str = "report train_data to remote_monitor error: {}, " \
                      "url:{}, headers:{}".format(resp.text, report_server, headers)
            logger.error(err_str)
            is_ok = False
            
    logger.error("report train_data time_used: %.4f" % (time.time() - st))
    return is_ok, err_str
