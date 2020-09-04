"""
读取 VOC 格式数据集公共方法
"""
import os
import xml.etree.ElementTree as ET
from util.logger import logger
from models.dataloader import data_loader_base
from util import tools
import time
import json

class VocLoader(data_loader_base.DataLoaderBase):
    def __init__(self, voc_path):
        """
        读取按照VOC格式生成的数据集
        :param voc_path: VOC格式的数据集目录，包含Annotations等目录
        """
        self._voc_path = voc_path

        self._path_anno = None  # Annotations
        self._path_imgs = None  # JPEGImages
        self._path_dataset_list = None  # ImageSets/Main
        self._init_path()
        super(VocLoader, self).__init__()

    def _init_path(self):
        self._path_anno = os.path.join(self._voc_path, 'Annotations')
        self._path_imgs = os.path.join(self._voc_path, 'JPEGImages')
        img_sets_path = os.path.join(self._voc_path, 'ImageSets')
        # 自定义数据里没有"Main"这一层，添加判断增强兼容性。
        self._path_dataset_list = os.path.join(img_sets_path, 'Main') if \
            os.path.isdir(os.path.join(img_sets_path, 'Main')) else \
            img_sets_path

    def _load_data_lists(self):
        """
        读取数据集编号列表，即VOC中 trainval.txt/test.txt等
        """
        for dataset in ['train', 'val']:
            dataset_list_file_path = os.path.join(self._path_dataset_list, dataset + '.txt')
            if os.path.exists(dataset_list_file_path):
                try:
                    self._dataset_lists[dataset] = [line.strip() for line in open(dataset_list_file_path)]
                except Exception as e:
                    logger.warning(dataset_list_file_path + repr(e))
        self._dataset_total = []
        for key in self._dataset_lists:
            self._dataset_total += self._dataset_lists[key]

    def _load_meta_data(self):
        """
        Annotation的每个xml文件包含了除了图片本身外全部信息，以此组织成数据集对象
        每个数据可属于 trainval.txt 或 test.txt等，但这里不加区分，读取全部xml文件，以编号为主键保存 self._meta_data_dict
        """
        logger.info("Preparing data ...")
        time_start = time.time()
        xml_data_list = sorted(os.listdir(self._path_anno))
        data_len = len(xml_data_list)
        idx = 0
        for annotation_file in xml_data_list:
            try:
                if annotation_file[-4:] != '.xml':
                    continue
                annotation_data = self._parse_xml_data(annotation_file)
                self._meta_data_dict[os.path.splitext(annotation_file)[0]] = annotation_data
                idx += 1
                if idx % 1000 == 0:
                    logger.info('Loading data: %d/%d' % (idx, data_len))
                tools.ratio_bar(num=idx, total=data_len, displayType='total')
            except Exception as e:
                logger.warning(annotation_file + '  ' + repr(e))
        print('')
        logger.info("Load meta data finished, total data num: %d, elapsed time: %d" %
            (len(self._meta_data_dict), int(time.time() - time_start)))

        # load label_map.json
        lablemap_path = self._voc_path + '/label_map.json'
        with open(lablemap_path, 'r') as f:
           tmp_list = json.loads(f.read())
           for k,v in tmp_list.items():
              self._try_add_class(v)

    def _parse_xml_data(self, annotation_file):
        """
        从xml文件获取数据信息，保存在 self._meta_data_dict[data_id] 中，格式为：
        {
            'data_id': 标准 VOC 数据集中 data_id 与 annotation_file 去掉扩展名后一致
            'folder': str，jpg 文件的目录，GOD系统中为完整路径
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
        :param annotation_file: xml文件路径
        :return: annotation_data
        """
        et = ET.parse(os.path.join(self._path_anno, annotation_file))
        element = et.getroot()
        element_obj_list = element.findall('object')
        element_folder = element.find('folder').text
        element_filename = element.find('filename').text
        element_width = int(float(element.find('size').find('width').text))
        element_height = int(float(element.find('size').find('height').text))
        annotation_data = {
            'data_id': annotation_file.replace('.xml', ''),
            # 虽然约定了VOC元数据 `<folder>`为图像所在绝对路径的目录，但给的用于测试代码的数据集却是标准VOC2012，
            # `<folder>`是个相对路径，为了在服务器上顺利测试，这里进行一个特判以自适应
            'folder': element_folder if element_folder[0] == '/' else self._path_imgs,
            'filename': element_filename,
            'width': element_width,
            'height': element_height,
            'bbox_list': []
        }
        for element_obj in element_obj_list:
            if element_obj.find('name') is None:
                logger.warning("The description of objects in %s may be broken" % annotation_file)
                continue
            label_name = element_obj.find('name').text
            self._try_add_class(label_name)
            obj_bbox = element_obj.find('bndbox')
            xmin = int(float(obj_bbox.find('xmin').text) + 1e-6)
            ymin = int(float(obj_bbox.find('ymin').text) + 1e-6)
            xmax = int(float(obj_bbox.find('xmax').text) + 1e-6)
            ymax = int(float(obj_bbox.find('ymax').text) + 1e-6)
            difficult = int(element_obj.find('difficult').text) == 1
            annotation_data['bbox_list'].append({
                'label': label_name,
                'xmin': xmin,
                'ymin': ymin,
                'xmax': xmax,
                'ymax': ymax,
                'difficult': difficult
            })
        return annotation_data

    # def get_meta(self, data_id):
    #     """
    #     暂停使用该方法，存在问题：不预先加载的话，在训练开始前无法知道分类总数和分类列表
    #     懒加载模式的get_meta，避免在数据集过大时，一次性加载整个 meta 列表需要很久（VOC2012图很多）
    #     :param data_id: str，数据编号
    #     :return: dict，元数据，如果不存在该编号则返回 None
    #     {
    #         'data_id': 标准数据集中 data_id 与 filename 去掉扩展名后一致，其他数据集理解为独立的唯一ID
    #         'folder': str，jpg 文件的目录，GOD中可能会是绝对路径
    #         'filename': str
    #         'width': str
    #         'height': str
    #         'bbox_list': [{
    #             'label': str, 类别名称，如 cat, dog 等
    #             'xmin': int
    #             'ymin': int
    #             'xmax': int
    #             'ymax': int
    #             'difficult': int, xml中的 difficult 字段
    #         }]
    #     }
    #     """
    #     annotation_file = self._meta_data_file_dict.get(data_id, None)
    #     if annotation_file is None:
    #         logger.error('wrong data_id: %s, annotation file not exist' % data_id)
    #         return None
    #         # raise RuntimeError('wrong data_id: %s, annotation file not exist' % data_id)
    #
    #     try:
    #         meta_data = self._parse_xml_data(annotation_file)
    #         if data_id not in self._meta_data_dict:
    #             self._meta_data_dict[data_id] = meta_data
    #         return meta_data
    #     except Exception as e:
    #         logger.error(('load annotation xml file failed: %s' % annotation_file) + ' ' + repr(e))
    #         return None
