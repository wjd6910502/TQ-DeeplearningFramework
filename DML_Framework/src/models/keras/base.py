
import json
import os
import time
import numpy as np
import tensorflow as tf
keras = tf.keras
K = keras.backend
import horovod.tensorflow.keras as hvd

from util.logger import logger
from . op import OpManager

class base_model(object):
  
  train_path = '' # 
  save_path = '' # auto param
  batch_size = 0 # auto param
  epoch = 10 # auto param
  hvd = None
  config = None
  model = None
  train_data_x = None 
  train_data_y = None
  valid_data_x = None
  valid_data_y = None
  learning_rate = 1e-3 # auto param
  local_rank = 0
  rank = 0
  iter = 1
  callback_list = []

  def __init__(self,args):  
    self.parse_params(args) 

     
  def parse_params(self,args):
    if args is None:
        return 

    self.train_path = args.train_path if args.train_path is not None else '' 
    self.save_path = args.save_path if args.save_path is not None else ''
    self.alg_type = args.alg_type if args.alg_type is not None else 'no type'
    self.param_initial = args.param_initial if args.param_initial is not None else 'uniform'
    self.op = args.op if args.op is not None else 'adgsum'
    self.batch_size = args.batch_size if args.batch_size is not None else '32'
    self.epochs = args.epochs if args.epochs is not None else 10
    self.lrate = args.lrate if args.lrate is not None else 0.01
    self.printparams()

  def hvd_init(self):
    # Horovod: initialize Horovod
    self.hvd = hvd
    self.config = tf.ConfigProto()
    self.hvd.init()

    K.clear_session()
    self.config.gpu_options.allow_growth = True
    self.config.gpu_options.visible_device_list = str(self.hvd.local_rank())
    K.set_session(tf.Session(config=self.config))
    
    # params
    self.local_rank = self.hvd.local_rank()
    self.hvd_size = self.hvd.size()
    self.rank = self.hvd.rank()
    #mgr
    self.opMgr = OpManager()

    #callback
    self.callback_list = [  self.hvd.callbacks.BroadcastGlobalVariablesCallback(0) ,
                            self.hvd.callbacks.MetricAverageCallback(),
                            self.hvd.callbacks.LearningRateWarmupCallback(warmup_epochs=5, verbose=1),
                            keras.callbacks.ReduceLROnPlateau(patience=10, verbose=1) 
                         ]
    
  #subclass inherit 
  def _pre_load_param(self):
    pass

  def _load_data(self):
    #rank0*batch_size + (rank0+1)*batch_size
    pass

  def _create_model(self):
    # define yourself model
    pass
     
  def _train_model(self):
    pass 
  
  def _save_model(self):
      
    pass
  
  def _evaluate(self):
    pass
  #subclass inherit 
 
  def run(self):
    self.hvd_init() 
    # inir_param
    self._pre_load_param()   
    # loaddata
    self._load_data()
    # create_model
    self._create_model()
    # train
    self._train_model()
    # evaluate
    self._evaluate()
    # save  
    self._save_model()
  
  def printparams(self):
    print('train_path = ',self.train_path) 
    print('save_path = ',self.save_path) 
    print('batch_size = ',self.batch_size)
    print('epoch = ',self.epoch) # auto param
    print('rank = ',self.rank)
    print('alg_type = ',self.alg_type)
    print('lrate = ',self.learning_rate)
    print('op = ',self.op) 

  def ConvDistribOp(self,op,learningrate):
    tmp_optimizer = self.opMgr.get(op)
    opt = tmp_optimizer(lr=learningrate*self.hvd_size)
    opt = self.hvd.DistributedOptimizer(opt)
    return opt 
    
  # decorate func
  def runtime(func):
    def wrapper(*args,**kwargs):
        start_time = time.time()
        get_str = func(*args, **kwargs)
        end_time = time.time()
        logger.info(str(func)+" cost_time: {}".format(end_time-start_time))
        return get_str
    return wrapper


