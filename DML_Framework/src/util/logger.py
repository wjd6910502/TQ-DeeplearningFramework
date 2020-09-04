import time
import json
import logging
import os
import sys

logger = logging.getLogger("")
formatter = logging.Formatter("[%(levelname)s][%(asctime)s]%(message)s", "%Y-%m-%d %H:%M:%S")

streamHandler = logging.StreamHandler(sys.stdout)
streamHandler.setFormatter(formatter)
logger.addHandler(streamHandler)

fileHandler = logging.FileHandler("./log/debug-python.log")
fileHandler.setFormatter(formatter)
logger.addHandler(fileHandler)

logger.setLevel(logging.DEBUG)


