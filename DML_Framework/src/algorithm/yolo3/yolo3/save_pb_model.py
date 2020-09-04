
import shutil
import os
import json
from collections import OrderedDict
from tensorflow import identity
from tensorflow.python import saved_model
from tensorflow.python.saved_model.signature_def_utils_impl import predict_signature_def
from util.logger import logger
from algorithm.yolo3.yolo3.model import my_yolo_boxes_and_scores
import tensorflow as tf
keras = tf.keras

def my_yolo_eval(args, anchors, num_classes):
    """Evaluate YOLO model on given input and return filtered boxes."""
    yolo_outputs0 = args[0]
    yolo_outputs1 = args[1]
    yolo_outputs2 = args[2]
    image_shape = args[3]
    num_layers = 3
    anchor_mask = [[6, 7, 8], [3, 4, 5], [0, 1, 2]] if num_layers == 3 else [[3, 4, 5], [1, 2, 3]]  # default setting
    input_shape = keras.backend.shape(yolo_outputs0)[1:3] * 32
    boxes = []
    box_scores = []
    box_confidence = []

    _boxes, _box_scores, _box_confidence = my_yolo_boxes_and_scores(yolo_outputs0,
                                                anchors[anchor_mask[0]], num_classes, input_shape, image_shape)
    boxes.append(_boxes)
    box_scores.append(_box_scores)
    box_confidence.append(_box_confidence)

    _boxes, _box_scores, _box_confidence = my_yolo_boxes_and_scores(yolo_outputs1,
                                                anchors[anchor_mask[1]], num_classes, input_shape, image_shape)
    boxes.append(_boxes)
    box_scores.append(_box_scores)
    box_confidence.append(_box_confidence)

    _boxes, _box_scores, _box_confidence = my_yolo_boxes_and_scores(yolo_outputs2,
                                                anchors[anchor_mask[2]], num_classes, input_shape, image_shape)
    boxes.append(_boxes)
    box_scores.append(_box_scores)
    box_confidence.append(_box_confidence)

    boxes = keras.backend.concatenate(boxes, axis=0)
    box_scores = keras.backend.concatenate(box_scores, axis=0)
    box_confidence = keras.backend.concatenate(box_confidence, axis=0)

    pre_nms_limit = tf.minimum(12000, tf.shape(box_confidence)[0])
    ix = tf.nn.top_k(box_confidence, pre_nms_limit, sorted=True,
                     name="top_anchors").indices
    box_scores = tf.gather(box_scores, ix)
    boxes = tf.gather(boxes, ix)
    box_confidence = tf.gather(box_confidence, ix)

    max_output_size = 200
    iou_threshold = 0.5
    score_threshold = 0.1

    # nms
    indices = tf.image.non_max_suppression(boxes, box_confidence, max_output_size, iou_threshold, score_threshold)  # 一维索引
    boxes = tf.gather(boxes, indices)  # (M,4)
    box_scores = tf.gather(box_scores, indices)  # 扩展到二维(M,1)

    boxes = pad_to_fixed_size(boxes, max_output_size)
    box_scores = pad_to_fixed_size(box_scores, max_output_size)

    boxes = keras.backend.expand_dims(boxes, 0)
    box_scores = keras.backend.expand_dims(box_scores, 0)

    return boxes, box_scores

def pad_to_fixed_size(input_tensor, fixed_size):
    """
    增加padding到固定尺寸,在第二维增加一个标志位,0-padding,1-非padding
    :param input_tensor: 二维张量
    :param fixed_size:
    :param negative_num: 负样本数量
    :return:
    """
    input_size = tf.shape(input_tensor)[0]
    # padding
    padding_size = tf.maximum(0, fixed_size - input_size)
    x = tf.pad(input_tensor, [[0, padding_size], [0, 0]], mode='CONSTANT', constant_values=0)
    return x

def data_process(agrs):
    image = agrs[0]
    image_shape = agrs[1]
    return image, image_shape

def h52pb(model, model_save_path):
    if os.path.isdir(model_save_path):
        shutil.rmtree(model_save_path)
    builder = saved_model.builder.SavedModelBuilder(model_save_path)

    signature = predict_signature_def(
        inputs={"input_image": model.inputs[0],
                "input_image_meta": model.inputs[1]},
        outputs={"box_scores": model.outputs[0],
                 "boxes": model.outputs[1]},
    )

    sess = keras.backend.get_session()
    builder.add_meta_graph_and_variables(sess=sess,
                                         tags=[saved_model.tag_constants.SERVING],
                                         signature_def_map={
                                             "serving_default": signature})
    builder.save()

#def _main():
#    cfg_path = 'evaluate_config.json'
#
#    if not os.path.exists(cfg_path):
#        logger.error("config file dose not exist: {}".format(cfg_path))
#        return False
#
#    with open(cfg_path, 'r') as file:
#        params = json.load(file, object_pairs_hook=OrderedDict)
#
#    save_path = "./models"
#    save_model(params, save_path)
#
#    return
#
#
#if __name__ == '__main__':
#    _main()


