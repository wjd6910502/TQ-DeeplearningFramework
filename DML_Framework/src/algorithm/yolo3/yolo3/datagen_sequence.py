
import tensorflow as tf
import numpy as np
from algorithm.yolo3.yolo3 import utils
from algorithm.yolo3.yolo3 import model as yolomodel
keras = tf.keras

#TODO: 结合keras框架，重写__get_item
class DataGenSequence(keras.utils.Sequence):
    def __init__(self, annotation_list, batch_size, input_size, step_num, anchors, num_classes):
        self.annotation_list = annotation_list
        self.item_num = len(self.annotation_list)
        self.batch_size = batch_size
        self.input_size = input_size
        self.step_num = step_num
        self.anchors = anchors
        self.num_classes = num_classes

    def __len__(self):
        return self.step_num

    def on_epoch_end(self):
        np.random.shuffle(self.annotation_list)

    def __getitem__(self, idx):
        img_data = []
        box_data = []
        for i in range(idx * self.batch_size, (idx + 1) * self.batch_size):
            ith = i % self.item_num
            img, box = utils.get_random_data(self.annotation_list[ith], self.input_size, random=True)
            img_data.append(img)
            box_data.append(box)
        img_data = np.array(img_data)
        box_data = np.array(box_data)
        y_true = yolomodel.preprocess_true_boxes(box_data, self.input_size, self.anchors, self.num_classes)
        return [img_data] + y_true, np.zeros(self.batch_size)
