
"""
在模型 fit_generator 中依靠 callback 对模型进行调整，比如修改一些层的 trainable，并重新 compile 使之生效
"""

import tensorflow as tf
keras = tf.keras
from util.logger import logger

class ModelModifyCallback(keras.callbacks.Callback):
    """
    用于在训练途中修改模型状态
    """
    def __init__(self, model_state_changing_epoch_list, learning_rate_list, trainable_changing_epoch, loss, opt ):
        self.model_state_changing_epoch_list = model_state_changing_epoch_list
        self.learning_rate_list = learning_rate_list
        self.trainable_changing_epoch = trainable_changing_epoch
        self.loss = loss
        self.opt = opt
        super(ModelModifyCallback, self).__init__()

    def on_epoch_begin(self, epoch, logs=None):
        for i in range(len(self.model_state_changing_epoch_list)):
            if epoch == self.model_state_changing_epoch_list[i]:
                if epoch == self.trainable_changing_epoch:
                    logger.info("changing trainable weights")
                    for l in range(len(self.model.layers)):
                        self.model.layers[l].trainable = True
                logger.info("recompiling model for training stage changing (%d stage)" % (i + 1))
                
                # Horovod: adjust learning rate based on number of GPUs.
                opt = keras.optimizers.Adam(lr=self.learning_rate_list[i])
                 
                # Horovod: add Horovod Distributed Optimizer.
                opt = self.opt(opt)    

                self.model.compile(optimizer=opt, loss=self.loss)
