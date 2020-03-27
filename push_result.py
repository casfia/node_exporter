#! /usr/bin/env python3
import os
import sched
import time
import subprocess
import configparser
import json
import requests
import logging

# 初始化sched模块的scheduler类
# 第一个参数是一个可以返回时间戳的函数，第二个参数可以在定时未到达之前阻塞。
schedule = sched.scheduler(time.time, time.sleep)
device_ip = ''
remote_url = ''
period = 0
program = ''
#初始化日志处理部分
logger = logging.getLogger(__name__)
logger.setLevel(level = logging.INFO)
handler = logging.FileHandler("/var/log/monior_push.log")
handler.setLevel(logging.INFO)
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)

# 被周期性调度触发的函数
def execute_command(cmd, inc):
    try:
        url = "http://localhost:9100/metrics"
        # 打开请求，获取对象
        response = requests.get(url)
        # 打印Http状态码
        # 读取服务器返回的数据,对HTTPResponse类型数据进行读取操作
        the_page = response.text
        json_str = json.dumps({'target_ip': device_ip,'program': program,'metrics_str': str(the_page),'period': int(period)})
        headers = {"Content-type": "application/json","Accept": "*/*"}
        response = requests.post(remote_url,data=json_str,headers=headers)
        logger.info('send a metric batch success.')
    except Exception as e:
        logger.info('send message error ',e)
        # 中文编码格式打印数据
    schedule.enter(inc, 0, execute_command, (cmd, inc))

def main(cmd, inc=60):
    # enter四个参数分别为：间隔事件、优先级（用于同时间到达的两个事件同时执行时定序）、被调用触发的函数，
    # 给该触发函数的参数（tuple形式）
    schedule.enter(0, 0, execute_command, (cmd,inc))
    schedule.run()

def get_ip():
    p = subprocess.Popen("hostname -I", shell=True, stdout=subprocess.PIPE)
    data = p.stdout.read() # 获取命令输出内容
    data = str(data,encoding = 'UTF-8') # 将输出内容编码成字符串
    ip_list = data.split(' ') # 用空格分隔输出内容得到包含所有IP的列表
    if "\n" in ip_list: # 发现有的系统版本输出结果最后会带一个换行符
        ip_list.remove("\n")
    return ip_list

if __name__ == '__main__':
    file = os.path.abspath(os.path.join(os.getcwd(),'config.ini'))
    cf = configparser.ConfigParser()
    cf.read(file,encoding='utf-8')
    remote_url = cf.get('remote','url')
    prefix = cf.get('local','ip_prefix')
    period = cf.get('local','period')
    program = cf.get('local','program')
    device_ips = get_ip()
    for ip_addr in device_ips:
        if ip_addr.startswith(prefix):
            device_ip = ip_addr
            break;
    if device_ip is '':
        sys.exit(1)
    pid=os.fork()
    if pid != 0:
        os._exit(0)
    else:
        main("",int(period))