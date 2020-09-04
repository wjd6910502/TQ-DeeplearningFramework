import os
import numpy as np
import tensorflow as tf
keras = tf.keras
from models.keras.base import base_model
from util.logger import logger
from algorithm.cnn.cnn_train import CnnTrain


class CNN(base_model):
    input_shape = (28, 28, 1)
    num_classes = 10

    def __init__(self, args):
        base_model.__init__(self, args)
        self.train_params = {}
        self.algorithm = None
        self.x_train, self.y_train, self.x_test, self.y_test = None, None, None, None

    def _pre_load_param(self):
        if self.save_path != "":
            self.train_params["model_path"] = self.save_path
        self.train_params["train_path"] = self.train_path
        self.train_params["param_initial"] = self.param_initial
        self.train_params["opt_name"] = self.op
        self.train_params["batch_size"] = int(self.batch_size)
        self.train_params["epochs"] = int(self.epochs)
        self.train_params["lr"] = float(self.lrate)
        logger.info("train_params:{}".format(self.train_params))

    def _load_data(self):
        data_path = self.train_params.get("train_path")
        if not os.path.exists(data_path):
            logger.error("data_path not exist: {}".format(data_path))
        path = os.path.join(data_path, "mnist.npz") if os.path.isdir(data_path) else data_path
        with np.load(path) as f:
            x_train, y_train = f['x_train'], f['y_train']
            x_test, y_test = f['x_test'], f['y_test']
        x_train = x_train.reshape(x_train.shape[0], self.input_shape[0], self.input_shape[1], self.input_shape[2])
        x_test = x_test.reshape(x_test.shape[0], self.input_shape[0], self.input_shape[1], self.input_shape[2])

        x_train = x_train.astype('float32')
        x_test = x_test.astype('float32')
        self.x_train = x_train/255
        self.x_test = x_test/255
        logger.info('x_train shape:{}'.format(x_train.shape))
        logger.info('x_test shape:{}'.format(x_test.shape))

        # Convert class vectors to binary class matrices
        self.y_train = keras.utils.to_categorical(y_train, self.num_classes)
        self.y_test = keras.utils.to_categorical(y_test, self.num_classes)

    def _train_model(self):
        self.algorithm = CnnTrain(self.input_shape, self.num_classes)
        self.algorithm.init(self.train_params)
        opt = self.ConvDistribOp(self.train_params.get("opt_name", "adadelta"), float(self.train_params["lr"]))
        self.algorithm.train(opt, self.callback_list, self.x_train, self.y_train, self.x_test, self.y_test)

    def _evaluate_model(self):
        eval_result = self.algorithm.evaluate_model(self.train_params, self.data_loader)

    def _save_model(self):
        self.algorithm.save_model()

