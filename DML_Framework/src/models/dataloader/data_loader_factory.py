from util.logger import logger
from models.dataloader import god_type_loader, voc_type_loader, txt_type_loader
from util import tools
from enum import Enum
from manager.s3manager import S3mgr  
from manager.filemanager import Filemanager 

class DataType(Enum):
    FLAGS_DATA_TYPE_VOC_SP   = 0
    FLAGS_DATA_TYPE_VOC      = 1
    FLAGS_DATA_TYPE_TXT      = 2
    FLAGS_DATA_TYPE_OTHER    = 3

TMPDIR = '/tmpdata'
TMPFILENAME = '/tmpdata/tmpdata.tar.gz'

#filename = slice_57640ece-c678-11ea-a45d-7af6424c9416_1595841231_1
def get_dataset(filename):    
    Filemanager.createfile(TMPDIR)  
    tarname = filename + ".tar.gz"
    s3 = S3mgr()
    s3.download_file("",tarname,TMPFILENAME) 
    Filemanager.untar(TMPFILENAME,TMPDIR)
    return TMPDIR + '/' + filename

def get_data_loader(args):
    """
    :param data_type: 传入参数的数据集类型
    :param kwargs:
    """
    if args['data_type'] == DataType.FLAGS_DATA_TYPE_VOC_SP.value:
        logger.info('VOC Splite type data, using VocLoader')
        voc_path = get_dataset(args['data_set'])
        return voc_type_loader.VocLoader(voc_path=voc_path)

    elif args['data_type'] == DataType.FLAGS_DATA_TYPE_VOC.value:
        logger.info('VOC type data, using VocLoader')
        voc_path = args['data_set']
        return voc_type_loader.VocLoader(voc_path=voc_path)

    elif args['data_type'] == DataType.FLAGS_DATA_TYPE_TXT.value:
        logger.info('TXT type data, using TxtLoader')
        return txt_type_loader.TxtLoader( label_path=args['label_path'], anno_path=args['anno_path'], train_ratio=args.get('trn_rate', 0.8))

    else:
        logger.error('Not supported data_type: %s.' % args['data_type'])
        exit()

#get_dataset('slice_57640ece-c678-11ea-a45d-7af6424c9416_1595841231_1.tar.gz')

     
