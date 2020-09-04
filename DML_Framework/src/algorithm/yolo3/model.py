import json
import os
import numpy as np
import tensorflow as tf
import time
keras = tf.keras
import traceback
#import horovod.tensorflow.keras as hvd
from tensorflow import identity
from util.logger import logger
from algorithm.yolo3.yolo3.kmeans import get_anchors
from algorithm.yolo3.yolo3.model import preprocess_true_boxes, yolo_body, yolo_loss
from algorithm.yolo3.yolo3.model_modify_callback import ModelModifyCallback
from algorithm.yolo3.yolo3 import datagen_sequence
from algorithm.yolo3.yolo3.save_pb_model import my_yolo_eval, h52pb
from models.keras.base import base_model
# TODO: factory --> manager
from models.dataloader import data_loader_factory
K = keras.backend 

class YOLO3(base_model):
    # base interface
    def __init__(self,args):
       super().__init__(args)
       # train params
       # basemodel params
       self.__train_path = self.train_path
       self.__batch_size = int(self.batch_size) 
       self.__max_epochs = int(self.epochs)
       self.__learning_rate = float(self.lrate)
       self.__load_pretrained_weights = True
       #TODO:
       self.__pretrained_weights_path = os.path.join('/data1/tq/pretrained','darknet53_weights.h5')
       self.train_steps = -1 
       self.valid_steps = -1

       # model params
       self.__anchors = None
       self.__class_names = list()
       self.__input_size = (416, 416)
       self.__num_anchors = 9
       self.__num_classes = 0
       self.__save_weights_name = 'yolov3_weights_final.h5'
       self.__data_loader = None
    
    def _pre_load_param(self):
        pass
        
    # base interface
    def _load_data(self):
        
        # data-type param setting
        d_args = {}
        logger.info("@@@@@@@@@@@@@@@@@@rank = {} ".format(self.hvd.rank()))
        logger.info("@@@@@@@@@@@@@@@@@@local_rank = {}".format(self.hvd.local_rank()))
        if self.train_path.startswith("slice_"): 
            d_args['data_type'] = 0
            d_args['data_set'] = self.train_path + str(self.hvd.rank())
        else:
            d_args['data_type'] = 1
            d_args['data_set'] = self.train_path
        self.__data_loader = data_loader_factory.get_data_loader(d_args) 
        
        self.__train_path = "/tmpdata/" + self.train_path + str(self.hvd.rank())
         
        self.__class_names = self.__data_loader.class_list
        self.__input_size = ( 416 ,416 )
        
        # 生成txt文件
        self._generate_data_list()

        self.__anchors = get_anchors(os.path.join(self.__train_path, 'train_data_list.txt')) 
        self.__num_anchors = len(self.__anchors)
        #self.__num_classes = len(self.__class_names)
        self.__num_classes = 27
        # 生成txt文件
        # self._generate_data_list()
        
        logger.info('****************model parameter****************')
        logger.info('anchors: {}'.format(np.array(self.__anchors).reshape(1, -1)[0].tolist()))
        logger.info('class names: {}'.format(self.__class_names))
        logger.info('input height: {}'.format(self.__input_size[0]))
        logger.info('input width: {}'.format(self.__input_size[1]))
        logger.info('number of anchors: {}'.format(self.__num_anchors))
        logger.info('number of classes: {}'.format(self.__num_classes))
        logger.info('************************************************')

    def _generate_data_list(self):
        #### TODO: 放在dataload里面，聚类
        def gen(ids, file):
            with open(file, 'w') as f:
                for img_id in ids:
                    #logger.info("@@@@@@@@img_id = {}".format(img_id))
                    meta_data = self.__data_loader.get_meta(img_id)
                    if meta_data is None:
                        #logger.info("%%%%%%%%%%%%%%%%%%%%%img_id = {} not exist".format(img_id))
                        continue
                    folder = meta_data['folder']
                    filename = meta_data['filename']
                    f.write(os.path.join(folder, filename))
                    
                    logger.info("meta_data = {}".format(meta_data['bbox_list']))
                     
                    for box in meta_data['bbox_list']:
                        difficult = box['difficult']
                        cls = box['label']

                        if cls not in self.__class_names or int(difficult) == 1:
                            continue
                        cls_id = self.__class_names.index(cls)
                        b = (box['xmin'], box['ymin'], box['xmax'], box['ymax'])
                        f.write(" " + ",".join([str(a) for a in b]) + ',' + str(cls_id))
                    f.write('\n')

        # generate train data list
        train_img_ids = self.__data_loader.get_dataset_id_list(dataset_name='train')
        train_file = os.path.join(self.__train_path, 'train_data_list.txt')
        gen(train_img_ids, train_file)
        logger.info('generate train file')
        # generate valid data list
        valid_img_ids = self.__data_loader.get_dataset_id_list(dataset_name='val')
        valid_file = os.path.join(self.__train_path, 'valid_data_list.txt')
        gen(valid_img_ids, valid_file)
        logger.info('generate valid file')
        return True

    def create_model(self, freeze_body=1):
        input_tensor = keras.layers.Input(shape=(None, None, 3))
        height, width = self.__input_size

        y_true = [keras.layers.Input(shape=(height // {0: 32, 1: 16, 2: 8}[l], width // {0: 32, 1: 16, 2: 8}[l],
                                            self.__num_anchors // 3, self.__num_classes + 5)) for l in range(3)]

        model_body = yolo_body(input_tensor, self.__num_anchors // 3, self.__num_classes)
        logger.info('create YOLOv3 model with {} anchors and {} classes'.format(self.__num_anchors, self.__num_classes))

        if self.__load_pretrained_weights is True:
            model_body.load_weights(self.__pretrained_weights_path, by_name=True)
            logger.info('load pretrained weights: {}'.format(self.__pretrained_weights_path))

            if freeze_body in [1, 2]:
                # Freeze darknet53 body or freeze all but 3 output layers.
                num = (185, len(model_body.layers) - 3)[freeze_body - 1]
                for i in range(num):
                    model_body.layers[i].trainable = False
                logger.info('freeze the first {} layers of total {} layers'.format(num, len(model_body.layers)))

        model_loss = keras.layers.Lambda(yolo_loss, output_shape=(1,), name='yolo_loss',
                                         arguments={'anchors': self.__anchors,
                                                    'num_classes': self.__num_classes,
                                                    'ignore_thresh': 0.5})(model_body.output + y_true)  # list合并
        model = keras.models.Model([model_body.input] + y_true, model_loss)

        return model

    # base interface 
    def _create_model(self):
        
        # get Horovod size.
        hvd_size = self.hvd_size

        # create the training model
        self.model = self.create_model(freeze_body=1)
 
        with open(os.path.join(self.__train_path, 'train_data_list.txt')) as f:
            train_list = f.readlines()
        np.random.seed(10101)
        np.random.shuffle(train_list)
        np.random.seed(None)
        num_train = len(train_list)
        # train_data = self._generate_data(train_list)

        # load valid data
        with open(os.path.join(self.__train_path, 'valid_data_list.txt')) as f:
            valid_list = f.readlines()
        np.random.seed(10101)
        np.random.shuffle(valid_list)
        np.random.seed(None)
        num_valid = len(valid_list)
        # valid_data = self._generate_data(valid_list)

        logger.info('train on {} samples, valid on {} samples, with batch size {}'.format(num_train, num_valid,
                                                                                          self.__batch_size))
        # set steps
        self.train_steps = max(1, num_train // self.__batch_size) if self.train_steps == -1 else self.train_steps
        self.valid_steps = max(1, num_valid // self.__batch_size) if self.valid_steps == -1 else self.valid_steps

        self.loss = {'yolo_loss': lambda y_true, y_pred: y_pred}
        self.first_stage_epochs = int(np.floor(self.__max_epochs * 0.50))
        self.second_stage_epochs = int(np.floor(self.__max_epochs * 0.75))
        self.third_stage_epochs = int(np.floor(self.__max_epochs * 0.875))

        # Horovod: adjust learning rate based on number of GPUs.
        opt = keras.optimizers.Adam(lr=self.__learning_rate * hvd_size)
        
        # Horovod: add Horovod Distributed Optimizer.
        opt = self.hvd.DistributedOptimizer(opt)    

        self.model.compile(
            optimizer=opt,
            loss=self.loss
        )

        self.data_gen_train = datagen_sequence.DataGenSequence(
            annotation_list=train_list,
            batch_size=self.__batch_size,
            input_size=self.__input_size,
            step_num=self.train_steps,
            anchors=self.__anchors,
            num_classes=self.__num_classes
        )

        self.data_gen_val = datagen_sequence.DataGenSequence(
            annotation_list=valid_list,
            batch_size=self.__batch_size,
            input_size=self.__input_size,
            step_num=self.valid_steps,
            anchors=self.__anchors,
            num_classes=self.__num_classes
        )
      
    # base interface
    def _train_model(self):
        start = time.time()
        callbacks = [] 
        callbacks += [
            # Horovod: broadcast initial variable states from rank 0 to all other processes.
            # This is necessary to ensure consistent initialization of all workers when
            # training is started with random weights or restored from a checkpoint.
            self.hvd.callbacks.BroadcastGlobalVariablesCallback(0),
            self.hvd.callbacks.MetricAverageCallback(),
            self.hvd.callbacks.LearningRateWarmupCallback(warmup_epochs=5, verbose=1),
            keras.callbacks.ReduceLROnPlateau(patience=10, verbose=1),            
        ]
        
        callbacks.append(ModelModifyCallback(
           model_state_changing_epoch_list=[
               self.first_stage_epochs,
               self.second_stage_epochs,
               self.third_stage_epochs,
           ],
           learning_rate_list=[
               self.__learning_rate * 0.1 * self.hvd_size,
               self.__learning_rate * 0.01 * self.hvd_size,
               self.__learning_rate * 0.001 * self.hvd_size,

           ],
           trainable_changing_epoch=self.first_stage_epochs,
           loss=self.loss,
           opt = self.hvd.DistributedOptimizer 
       ))

        try:
            self.model.fit_generator(
                generator=self.data_gen_train,
                steps_per_epoch=self.train_steps,
                validation_data=self.data_gen_val,
                validation_steps=self.valid_steps,
                epochs=self.__max_epochs,
                initial_epoch=0,
                callbacks=callbacks,
                # use_multiprocessing=True,
                # workers=1
            )
        except Exception as err:
            logger.error("train errr: {}, traceback:{}".format(err, traceback.format_exc()))
        
        end = time.time()
        training_time = end - start
        logger.info('training model spent %d',training_time)
        if self.hvd.rank() == 0:
            self.model.save_weights(os.path.join(self.save_path, self.__save_weights_name))
        
        logger.info('training complete')

        return True
    
    def _save_model(self):
        pass
        #self.save_model(params,path) 

    def _evaluate(self): 
        pass    
    def save_config(self):
        # self._save_predict_config()
        evaluate_cfg = self._save_evaluate_config()

        return evaluate_cfg
    
    def _save_predict_config(self):
        cfg = dict()

        cfg['model'] = dict()
        cfg['model']['anchors'] = np.array(self.__anchors).reshape(1, -1)[0].tolist()
        cfg['model']['class_names'] = self.__class_names
        cfg['model']['input_width'] = self.__input_size[0]
        cfg['model']['input_height'] = self.__input_size[1]

        cfg['train'] = dict()
        cfg['train']['save_folder'] = self.__train_path
        cfg['train']['batch_size'] = self.__batch_size
        cfg['train']['max_epochs'] = self.__max_epochs
        cfg['train']['learning_rate'] = self.__learning_rate
        cfg['train']['load_pretrained_weights'] = self.__load_pretrained_weights
        cfg['train']['pretrained_weights_path'] = self.__pretrained_weights_path

        cfg['predict'] = dict()
        cfg['predict']['weights_path'] = os.path.join(self.__train_path, self.__save_weights_name)
        cfg['predict']['iou_threshold'] = 0.45
        cfg['predict']['score_threshold'] = 0.50
        cfg['predict']['save_txt'] = True
        cfg['predict']['save_image'] = True

        with open('predict_config.json', 'w') as file:
            json.dump(cfg, file, indent=4)

        return cfg

    def _save_evaluate_config(self):
        cfg = dict()

        cfg['model'] = dict()
        cfg['model']['anchors'] = np.array(self.__anchors).reshape(1, -1)[0].tolist()
        cfg['model']['class_names'] = self.__class_names
        cfg['model']['input_width'] = self.__input_size[0]
        cfg['model']['input_height'] = self.__input_size[1]

        cfg['train'] = dict()
        cfg['train']['save_folder'] = self.__train_path
        cfg['train']['batch_size'] = self.__batch_size
        cfg['train']['max_epochs'] = self.__max_epochs
        cfg['train']['learning_rate'] = self.__learning_rate
        cfg['train']['load_pretrained_weights'] = self.__load_pretrained_weights
        cfg['train']['pretrained_weights_path'] = self.__pretrained_weights_path

        cfg['evaluate'] = dict()
        cfg['evaluate']['weights_path'] = os.path.join(self.__train_path, self.__save_weights_name)
        cfg['evaluate']['iou_threshold'] = 0.45
        cfg['evaluate']['score_threshold'] = 0.001
        cfg['evaluate']['save_txt'] = True
        cfg['evaluate']['save_image'] = False

        with open(os.path.join(self.__train_path, 'evaluate_config.json'), 'w') as file:
            json.dump(cfg, file, ensure_ascii=False, indent=4)

        return cfg

    def save_model(self, params, save_path):
        if self.hvd.rank() == 0:
            self._save_model1(params, save_path)

    def _save_model1(self, params, save_path):
        config = tf.ConfigProto()
        config.gpu_options.allow_growth = True
        keras.backend.set_session(tf.Session(config=config))

        params = param_preprocess.param_prepare(params, 'save')
        keras.backend.clear_session()

        anchors = np.array(params['model']['anchors']).reshape(-1, 2)

        num_anchors = 9
        num_classes = len(params['model']['class_names'])

        input_image = keras.layers.Input(shape=(None, None, 3), name='input_image')
        input_image_meta = keras.layers.Input(shape=(2,), name='input_image_meta')

        yolo_model = yolo_body(input_image, num_anchors // 3, num_classes)
        logger.info('load weights from {}'.format(params['evaluate']['weights_path']))
        yolo_model.load_weights(params['evaluate']['weights_path'])

        boxes, box_scores = keras.layers.Lambda(my_yolo_eval, output_shape=None, name='yolo_eval1',
                                                arguments={'anchors': anchors, 'num_classes': num_classes})(
            [yolo_model.output[0],
             yolo_model.output[1],
             yolo_model.output[2],
             input_image_meta])

        boxes = keras.layers.Lambda(lambda x: identity(x, name="boxes"))(boxes)
        box_scores = keras.layers.Lambda(lambda x: identity(x, name="box_scores"))(box_scores)

        model = keras.models.Model([input_image, input_image_meta], [box_scores, boxes])

        h52pb(model, save_path)
        logger.info('save pb model in {}'.format(save_path))



