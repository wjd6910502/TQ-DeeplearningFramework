
import importlib
import traceback
from util.logger import logger

class Algothmgr():

    def __init__(self, args):
        pass

    def create_algorithm(self, alg_type, args):
        try:
            module_name = 'algorithm.' + alg_type + '.model'
            class_name = getattr(importlib.import_module(module_name), alg_type.upper())
            return class_name(args)
        except Exception:
            logger.error("use alg_type:<{}> parse algorithm class exception: {}".format(alg_type,
                                                                                        traceback.format_exc()))
            return None
