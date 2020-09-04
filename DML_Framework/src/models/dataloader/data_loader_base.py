import os
import json
from util import tools
import pprint
from util.logger import logger


class DataLoaderBase:
    def __init__(self):
        """
        公共数据读取方法基类，包含数据集基本属性，设定了数据集结构化后的形式
        继承该类后需要提供所有公共变量与函数（不带开头下划线"_"）的内容
        """
        self._dataset_id_list_total = []    # 整个数据集 data_id 列表
        self._dataset_lists = {}    # trainval, train, val, test 的数据名字典，VOC中为 000xxxxx 的列表
        self._label_count = {}      # 整个数据集中每个label的数量，也是label是否已读到的标记，可为空

        self._meta_data_dict = {}   # xml提取信息的数据列表，以数据编号为键
        self.class_list = []
        ##################################################
        # 结构化数据集
        self._init_labels()         # 预定义类别列表已确保每次训练的label固定
        self._load_data_lists()
        self._load_meta_data()
        self.num_classes = len(self.class_list)
        self.class_name_to_id_dict = dict(list(zip(self.class_list, list(range(self.num_classes)))))
        self.class_id_to_name_dict = dict(list(zip(list(range(self.num_classes)), self.class_list)))
        ##################################################
        logger.info("class list: " + pprint.pformat(self.class_list))

    def __new__(cls, *args, **kwargs):
        # 单例，不重复实例化，减少读取元数据次数
        if not hasattr(cls, '_instance'):
            cls._instance = super().__new__(cls)
        return cls._instance

    def get_meta(self, data_id):
        """
        用数据编号获取元数据
        :param data_id: str，数据编号
        :return: dict，元数据，如果不存在该编号则返回 None
        {
            'data_id': 标准数据集中 data_id 与 filename 去掉扩展名后一致，其他数据集理解为独立的唯一ID
            'folder': str，jpg 文件的目录，GOD中可能会是绝对路径
            'filename': str
            'width': str
            'height': str
            'bbox_list': [{
                'label': str, 类别名称，如 cat, dog 等
                'xmin': int
                'ymin': int
                'xmax': int
                'ymax': int
                'difficult': int, xml中的 difficult 字段
            }]
        }
        """
        return self._meta_data_dict[data_id] if data_id in self._meta_data_dict else None

    def get_dataset_num(self, dataset_name):
        """
        获取 trainval、train、val、test等集合的数据个数
        :param dataset_name: str，trainval、train、val、test的其中一个，GOD中目前只有 train 和 val
        :return: int，如果不存在该集合则返回 None
        """
        return len(self._dataset_lists[dataset_name]) if dataset_name in self._dataset_lists else None

    def get_dataset_id_list(self, dataset_name):
        """
        获取集合列表，即 train、val 等其中之一的数据id列表
        :param dataset_name: str，trainval、train、val、test的其中一个，GOD中目前只有 train 和 val
        :return: list，data_id 列表，如参数错误返回 None
        """
        return self._dataset_lists[dataset_name] if dataset_name in self._dataset_lists else None

    def get_img(self, data_id):
        """
        根据编号获取对应的图片，BGR三通道[0,255]uint8格式
        :param data_id:
        :return: ndarray，如果 data_id 不存在或者图像文件不存在则返回 None
        """
        meta = self.get_meta(data_id)
        if meta is None:
            return None
        file_path = os.path.join(meta['folder'], meta['filename'])
        if not os.path.exists(file_path):
            return None
        return tools.imread(os.path.join(meta['folder'], meta['filename']))

    def get_groundtruth_files(self, groundtruth_file_dir, dataset_name):
        """
        把Annotation的数据转换为公共eval所需的 ground truth 格式，用来与预测数据计算mAP
        公共eval见 ai/detection/evaluate/voc_evaluation
        """
        self.dataset_id_list = self.get_dataset_id_list(dataset_name)
        for data_id in self.dataset_id_list:
            meta_data = self.get_meta(data_id)
            if meta_data is None:
                continue
            txt_name = data_id + '.txt'
            txt_path = os.path.join(groundtruth_file_dir, txt_name)
            with open(txt_path, 'w') as txt_file:
                for box in meta_data['bbox_list']:
                    difficult = box['difficult']
                    cls = box['label']
                    if cls not in self.class_name_to_id_dict:
                        continue
                    xmin = box['xmin']
                    ymin = box['ymin']
                    xmax = box['xmax']
                    ymax = box['ymax']
                    if int(difficult) == 1:
                        txt_file.write('%s %d %d %d %d difficult\n' % (cls, xmin, ymin, xmax, ymax))
                    else:
                        txt_file.write('%s %d %d %d %d\n' % (cls, xmin, ymin, xmax, ymax))

    def save_label_map(self, file_path):
        """
        保存 id->label的映射，格式为 {"id": "label_name"}
        :param file_path: 保存文件路径
        :return:
        """
        with open(file_path, "w") as f:
            json.dump(self.class_id_to_name_dict, f, ensure_ascii=False, indent=4)

    def _init_labels(self):
        """
        初始化 class 列表，也决定了 class name -> class id 的映射。
        重载该方法和 _parse_xml_data 可自定义 class name 的顺序
        :return: pass
        """
        # pass
        # # TODO: 检查算法各个环节，确认是否可去掉预设的背景类（设置此类涉及frcnn预测规则以及兼容无标记数据集问题）
        # self.class_list = ['__background__']
        # self._label_count = {'__background__': 0}
        self.class_list = []
        self._label_count = {}

    def _load_data_lists(self):
        """
        读取数据集编号列表，trainval、train、val、test 等
        """
        raise NotImplementedError

    def _load_meta_data(self):
        """
        读取整个数据集meta信息，即图像路径、标注信息等
        """
        raise NotImplementedError

    def _try_add_class(self, label_name):
        """
        添加类别
        :param label_name: str，类别名称
        """
        if label_name not in self._label_count:
            self._label_count[label_name] = 1
            self.class_list.append(label_name)
        else:
            self._label_count[label_name] += 1
