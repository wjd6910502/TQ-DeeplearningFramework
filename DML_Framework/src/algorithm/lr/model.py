'''
 line 1 is x ,line 2 is y 
 data=[
    [0.067732,3.176513],[0.427810,3.816464],[0.995731,4.550095],[0.738336,4.256571],[0.981083,4.560815],
    [0.526171,3.929515],[0.378887,3.526170],[0.033859,3.156393],[0.132791,3.110301],[0.138306,3.149813],
    [0.247809,3.476346],[0.648270,4.119688],[0.731209,4.282233],[0.236833,3.486582],[0.969788,4.655492],
    [0.607492,3.965162],[0.358622,3.514900],[0.147846,3.125947],[0.637820,4.094115],[0.230372,3.476039],
    [0.070237,3.210610],[0.067154,3.190612],[0.925577,4.631504],[0.717733,4.295890],[0.015371,3.085028],
    [0.335070,3.448080],[0.040486,3.167440],[0.212575,3.364266],[0.617218,3.993482],[0.541196,3.891471]
 ]
 # gen x,y matrix
 dataMat = np.array(data)
 X = dataMat[:,0:1]   # val x
 y = dataMat[:,1]   # val y
'''


import numpy as np
from keras.models import Sequential
from keras.layers import Dense
from models.keras.base import base_model
from util.logger import logger

data=[
    [0.067732,3.176513],[0.427810,3.816464],[0.995731,4.550095],[0.738336,4.256571],[0.981083,4.560815],
    [0.526171,3.929515],[0.378887,3.526170],[0.033859,3.156393],[0.132791,3.110301],[0.138306,3.149813],
    [0.247809,3.476346],[0.648270,4.119688],[0.731209,4.282233],[0.236833,3.486582],[0.969788,4.655492],
    [0.607492,3.965162],[0.358622,3.514900],[0.147846,3.125947],[0.637820,4.094115],[0.230372,3.476039],
    [0.070237,3.210610],[0.067154,3.190612],[0.925577,4.631504],[0.717733,4.295890],[0.015371,3.085028],
    [0.335070,3.448080],[0.040486,3.167440],[0.212575,3.364266],[0.617218,3.993482],[0.541196,3.891471]
]


class LR(base_model):
    def __init__(self,args):
       super().__init__(args)

    def _load_data(self):
       dataMat = np.array(data)
       self.train_data_x = dataMat[:,0:1]   # val x
       self.train_data_y = dataMat[:,1]   # val y

    def _create_model(self):
       self.model = Sequential()
       self.model.add(Dense(input_dim=1, units=1))
       self.model.compile(loss='mse', optimizer='sgd')
                
    def _train_model(self):
       logger.info('Training -----------')
       for step in range(5000):
          cost = self.model.train_on_batch(self.train_data_x, self.train_data_y)
          if step % 50 == 0:
             logger.info("After %d trainings, the cost: %f" % (step, cost))
        
    def evaluate_model(self):
       logger.info('\nTesting ------------')
       cost = self.model.evaluate(self.train_data_x, self.train_data_y, batch_size=40)
       logger.info('test cost:', cost)
       W, b = model.layers[0].get_weights()
       logger.info('Weights=', W, '\nbiases=', b)
    
    def perict_model(self):
       pass

    
