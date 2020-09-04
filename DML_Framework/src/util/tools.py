import sys
import pickle
import cv2
import json
from PIL import Image
from util.logger import logger

# 一些可复用的简化代码或方便灵活修改实现方式的的小工具，如异常信息处理、保存持久化数据、绘制进度条等


def manage_exception(e):
    """
    自定义处理异常信息的方式
    :param e: `except Exception as e` 这类地方的异常信息
    :return: pass
    """
    logger.error(repr(e))


def save_pickle(py_var, file):
    """
    将变量保存为pickle
    :param var: 要保存的python变量
    :param file: pickle文件路径，xxx.pkl
    :return: pass
    """
    with open(file, 'wb') as output:
        pickle.dump(py_var, output)


def load_pickle(file):
    """
    从pkl文件读取python变量
    :param file: pkl文件
    :return: python变量 var
    """
    with open(file, 'rb') as pkl_file:
        py_var = pickle.load(pkl_file)
    return py_var


def ratio_bar(num, total, info='', displayType='total', bar_length=30):
    """
    自定义的百分比进度条
    :param num: 当前进度数
    :param total: 总数
    :param info: 显示在进度条外的额外信息
    :param displayType: 'percentage'百分比显示，或'total'按数量显示
    :param bar_length: 进度条显示宽度
    :return: pass
    """
    rate = float(num) / total
    rate_num = int(rate * bar_length)
    if rate_num < bar_length:
        r = '\r[%s>%s]' % ("=" * rate_num, "." * (bar_length - rate_num - 1))
    else:
        r = '\r[%s%s]' % ("=" * rate_num, "." * (bar_length - rate_num - 1))
    if displayType == 'percentage':
        r += '%d%%' % (rate_num)
    elif displayType == 'total':
        r += '%d/%d' % (num, total)
    r += ' ' + info
    sys.stdout.write(r)
    sys.stdout.flush()


def imread(file, channels=3):
    """
    封装图像读取，方便以后若更改图像库，只改该接口即可
    :param file: 图像绝对路径
    :param channels: 以彩色还是灰度读取
    :return: ndarray
    """
    if channels == 3:
        return cv2.imread(file, cv2.IMREAD_COLOR)
    return cv2.imread(file, cv2.IMREAD_GRAYSCALE)


def imsave(file, img):
    """
    封装图像保存，方便以后若更改图像库，只改该接口即可
    :param file: 图像绝对路径
    :param img: 图像 ndarray
    """
    cv2.imwrite(file, img)


def imsize(file):
    """
    利用 Pillow 的懒加载，获取图像尺寸而不将图片完整加载到内存
    :param file: 图像绝对路径
    :return: tuple, (height, width)
    """
    try:
        width, height = Image.open(file).size
        return height, width
    except:
        return None


def split_path(path):
    if sys.platform.startswith("win"):
        path.replace('\\', '/')
    ret = []
    for seg in path.split('/'):
        if seg != '':
            ret.append(seg)
    return ret


def load_json(json_file):
    with open(json_file, 'r') as f:
        return json.load(f)


def save_json(json_file, python_var, indent=None):
    with open(json_file, "w") as f:
        json.dump(python_var, f, ensure_ascii=False, indent=indent)
