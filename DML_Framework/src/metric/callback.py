
import time
import tensorflow as tf
keras = tf.keras
from metric.report_train_data import train_data_report
from util.logger import logger


class RemoteMonitor(keras.callbacks.Callback):
    def __init__(self, is_send=False, count_mode="samples"):
        super(RemoteMonitor, self).__init__()
        if count_mode == 'samples':
            self.use_steps = False
        elif count_mode == 'steps':
            self.use_steps = True

        self.is_send = is_send
        self.train_start_time = None
        self.epoch_start_time = None
        self.batch_start_time = None
        self.batch_total_time = None
        self.batch_report_interval_start_time = None

    def on_train_begin(self, logs=None):
        self.verbose = self.params['verbose']
        self.epochs = self.params['epochs']
        self.train_start_time = time.time()

    def on_epoch_begin(self, epoch, logs=None):
        if self.use_steps:
            self.target = self.params['steps']
        else:
            self.target = self.params['samples']

        self.epoch_start_time = time.time()
        self.seen = 0

        self.batch_total_time = None
        self.batch_start_time = None
        self.batch_report_interval_start_time = time.time()

    def on_batch_begin(self, batch, logs=None):
        if self.seen < self.target:
            self.log_values = []
        self.batch_start_time = time.time()
        self.batch_total_time = time.time()

    def on_batch_end(self, batch, logs=None):
        batch_size = logs.get('size', 0)
        num_steps = logs.get('num_steps', 1)
        if self.use_steps:
            self.seen += num_steps
        else:
            self.seen += batch_size * num_steps

        if time.time() - self.batch_report_interval_start_time > 30*1:  # report per minute
            time_used_per_batch = "%.3f" % ((time.time() - self.epoch_start_time) / self.seen)
            data = {"current_batch": self.seen, "batch_steps": self.target, "time_used_per_batch": time_used_per_batch}
            self.try_report(data)
            self.batch_report_interval_start_time = time.time()

    def on_epoch_end(self, epoch, logs=None):
        logs = logs or {}
        send = {}
        send['current_epoch'] = epoch
        for k, v in logs.items():
            send[k] = str(v)
        send["epochs"] = self.epochs
        send["time_used_per_epoch"] = "%.3f" % (time.time() - self.epoch_start_time)
        self.try_report(send)

    def on_train_end(self, logs=None):
        data = {"train_end": 1, "time_used_train": time.time() - self.train_start_time}
        self.try_report(data)

    def try_report(self, data):
        if not self.is_send:
            return
        try:
            is_ok, err_str = train_data_report(**data)
        except Exception as err:
            logger.warning('Exception: could not reach RemoteMonitor, exception:{}'.format(err))
        else:
            if not is_ok:
                logger.warning("report_train_data error:{}".format(err_str))
