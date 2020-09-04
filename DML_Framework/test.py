import os
import json


with open('test.json', 'r') as f:
    tmp_list = json.loads(f.read())
    for k,v in tmp_list.items():
        print("k =",k," value =",v)
