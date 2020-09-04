
import os
import numpy as np
import tensorflow as tf
keras = tf.keras

class OpManager(object):
    op_dict = {}

    def __init__(self):
        self.register()

    def register(self):
        self.op_dict['sgd'] = keras.optimizers.Adadelta
        self.op_dict['rmsprop'] = keras.optimizers.RMSprop
        self.op_dict['adam'] = keras.optimizers.Adam
        self.op_dict['adadelta'] = keras.optimizers.Adadelta
        self.op_dict['adagrad'] = keras.optimizers.Adagrad
        self.op_dict['adamax'] = keras.optimizers.Adamax
        self.op_dict['nadam'] = keras.optimizers.Nadam
        #self.op_dict['ftrl'] = keras.optimizers.Ftrl
        #op_dict['fm'] = youself define optimize  

    def get(self,op):
        if self.op_dict.get(op) is None:
            return self.op_dict.get('adagrad') # return default

        return self.op_dict.get(op)


