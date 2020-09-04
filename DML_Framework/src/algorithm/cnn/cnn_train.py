import os
import tensorflow as tf
import time

keras = tf.keras
import horovod.tensorflow.keras as hvd

from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Dense, Dropout, Flatten
from tensorflow.keras.layers import Conv2D, MaxPooling2D
from metric.callback import MetricMonitor
K = keras.backend
from util.logger import logger


class CnnTrain(object):
    save_weights_name = "cnn"

    def __init__(self, input_shape, num_classes):
        self.input_shape = input_shape
        self.num_classes = num_classes
        self.params = None

        self.model = None

    def init(self, params,):
        self.params = params
        return True

    def train(self, opt, callbacks, x_train, y_train, x_test, y_test):
        start_time = time.time()
        # create the training model
        logger.info("input_shape:{}, output_shape:{}".format(self.input_shape, self.num_classes))
        self.model = self._create_model(self.input_shape, self.num_classes)

        self.model.compile(loss=keras.losses.categorical_crossentropy,
                           optimizer=opt,
                           metrics=['accuracy'])

        # Horovod: save checkpoints only on worker 0 to prevent other workers from corrupting them.
        if hvd.rank() == 0:
            save_path = "./model_save" if self.params.get("model_path") is None else self.params.get("model_path")
            if not os.path.exists(save_path):
                os.makedirs(save_path)
            callbacks.append(keras.callbacks.ModelCheckpoint(os.path.join(save_path,
                                                                          '%s-{epoch}.h5' % self.save_weights_name)))

        # add remote monitor callback
        callbacks.append(MetricMonitor(is_send=hvd.rank() == 0))
        batch_size = int(self.params.get("batch_size", 8))
        epochs = int(self.params.get("epochs", 20))
        steps_per_epoch = len(x_train) // batch_size
        logger.info("run epoch:{}, steps_per_epoch:{}, batch_size:{}".format(epochs, steps_per_epoch, batch_size))
        self.model.fit(x_train, y_train,
                       batch_size=batch_size,
                       callbacks=callbacks,
                       epochs=epochs,
                       verbose=1 if hvd.rank() == 0 else 0,
                       validation_data=(x_test, y_test))

        training_time = time.time() - start_time
        logger.info('training model spent %d', training_time)
        logger.info("evaluate ....")
        score = self.model.evaluate(x_test, y_test, verbose=0)
        logger.info('Test loss:{}'.format(score[0]))
        logger.info('Test accuracy:{}'.format(score[1]))

        logger.info('training complete')
        return True

    def add_conv(self, model, layer_num, block_num=1, filter_num=64, kernel_size=(3, 3)):
        for i in range(0, layer_num):

            model.add(Conv2D(filter_num, kernel_size, activation='relu', padding="same",
                             name="conv%s_%d" % (block_num, i)))
        return model

    def _create_model(self, input_shape, num_classes):
        model = Sequential()
        model.add(Conv2D(64, kernel_size=(3, 3),
                         activation='relu',
                         input_shape=input_shape, name="conv1_1"))
        model = self.add_conv(model, 1, block_num=1, filter_num=64, kernel_size=(3, 3))
        model.add(MaxPooling2D(pool_size=(2, 2), padding="same"))

        model = self.add_conv(model, 2, block_num=2, filter_num=128, kernel_size=(3, 3))
        model.add(MaxPooling2D(pool_size=(2, 2), padding="same"))

        model = self.add_conv(model, 3, block_num=3, filter_num=256, kernel_size=(3, 3))
        model.add(MaxPooling2D(pool_size=(2, 2), padding="same"))

        model = self.add_conv(model, 3, block_num=4, filter_num=512, kernel_size=(3, 3))
        model.add(MaxPooling2D(pool_size=(2, 2), padding="same"))

        # model = self.add_conv(model, 3, block_num=5, filter_num=512, kernel_size=(3, 3))
        # model.add(MaxPooling2D(pool_size=(2, 2), padding="same"))

        model.add(Flatten())
        # model.add(Dense(4096, activation='relu'))
        # model.add(Dropout(0.5))
        model.add(Dense(4096, activation='relu'))
        model.add(Dropout(0.5))
        model.add(Dense(num_classes, activation='softmax'))
        logger.info("mode.summary:{}".format(model.summary()))
        return model

    def save_model(self):
        if hvd.rank() == 0:
            save_path = self.params.get("model_path")
            if not os.path.exists(save_path):
                os.makedirs(save_path)
            self.model.save_weights(os.path.join(save_path, '%s-final.h5' % self.save_weights_name))

