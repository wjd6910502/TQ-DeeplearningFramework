
"""
读取 每行一个图描述 格式数据集公共方法，例如：
/data/dataset/test/000001.jpg 122,187,136,201,7 71,209,85,223,4 78,7,100,29,2 366,244,394,272,3 344,383,372,411,3 238,243,294,299,8 24,233,136,345,8
"""
import os
import time
import numpy as np
from models.dataloader import data_loader_base
from util import tools
from util.logger import logger

class TxtLoader(data_loader_base.DataLoaderBase):
    def __init__(self, label_path, anno_path, train_ratio=0.8):
        """
        读取每行一个描述的数据集
        :param label_path: 一列的文本文件，从0开始计数的每个分类的名称
        :param anno_path: 数据描述文件，每行第一个是文件名，所在路径与anno_path一致，
                            后续每组为由 xmin,ymin,xmax,ymax,class_id 组成的标注框
        :param train_ratio: 作为训练集的比例，其余作为验证集
        """
        self._label_path = label_path
        self._anno_path = anno_path
        self._trn_rate = train_ratio
        self._annotation_lines = []     # 每行标注的字符串

        super(TxtLoader, self).__init__()

    def _init_labels(self):
        """
        初始化 class 列表，也决定了 class name -> class id 的映射。
        重载该方法和 _parse_xml_data 可自定义 class name 的顺序
        :return: pass
        """
        with open(self._label_path, 'r') as data:
            for name in data:
                class_name = name.strip('\n')
                self.class_list.append(class_name)
                self._label_count[class_name] = 0

    def _load_data_lists(self):
        """
        读取数据集编号列表，即VOC中 trainval.txt/test.txt等
        """
        with open(self._anno_path, 'r') as f:
            txt = f.readlines()
            self._annotation_lines = [line.strip() for line in txt if len(line.strip().split()[1:]) != 0]
        self._dataset_id_list_total = [str(_) for _ in range(1, len(self._annotation_lines) + 1)]

        train_num = int(len(self._dataset_id_list_total) * self._trn_rate)
        self._dataset_lists['train'] = self._dataset_id_list_total[:train_num]
        self._dataset_lists['val'] = self._dataset_id_list_total[train_num:]

    def _load_meta_data(self):
        """
        Annotation的每个xml文件包含了除了图片本身外全部信息，以此组织成数据集对象
        每个数据可属于 trainval.txt 或 test.txt等，但这里不加区分，读取全部xml文件，以编号为主键保存 self.meta_data_dict
        """
        logger.info("Preparing data ...")
        time_start = time.time()
        # 无法保证文件名是否重名，所以使用编号
        data_len = len(self._annotation_lines)
        for idx in range(data_len):
            if idx % 1000 == 0:
                logger.info('Loading data: %d/%d' % (idx, data_len))
            anno_line = self._annotation_lines[idx].split()
            bboxes = [list(map(int, box.split(','))) for box in anno_line[1:]]
            data_id = self._dataset_id_list_total[idx]
            meta_data = dict({
                'data_id': data_id,
                # 去掉开头的"/static"，参考 `doc/default_configs/parameter_global.example.json`的["mark"]["data"]
                'folder': os.path.dirname(self._anno_path),
                'filename': anno_line[0],
                'width': None,  # 在后文处理
                'height': None,
                'bbox_list': []
            })
            img_path = os.path.join(meta_data['folder'], meta_data['filename'])
            img_size = tools.imsize(img_path)
            if img_size is None:
                logger.error('cannot get image %s' % img_path)
                exit()
            assert img_size[0] > 32 and img_size[1] > 32    # 检验一下训练数据合法性，太小的话可能数据有问题
            meta_data['height'] = img_size[0]
            meta_data['width'] = img_size[1]
            for obj in bboxes:
                box = {
                    'label': self.class_list[obj[4]],
                    'xmin': obj[0],
                    'ymin': obj[1],
                    'xmax': obj[2],
                    'ymax': obj[3],
                    'difficult': 0
                }
                meta_data['bbox_list'].append(box)
                assert box['xmax'] > box['xmin'] >= 0 and box['ymax'] > box['ymin'] >= 0
                self._try_add_class(box['label'])
            self._meta_data_dict[data_id] = meta_data
        logger.info(
            "Load meta data finished, total data num: %d, elapsed time: %d" %
            (len(self._meta_data_dict), int(time.time() - time_start))
        )
