import ipaddress
import os
import json
import logging
import socket
import tempfile

import psutil
from apscheduler.schedulers.blocking import BlockingScheduler

from settings import settings
from cfdnsapi import *


def main():
    # 获取网卡IP地址
    ipv4, ipv6 = get_network_addresses(settings['network_interface'])
    logging.info("读取网络接口[%s]: 公网IPv4 [%s] 公网IPv6 [%s]", settings['network_interface'], ipv4, ipv6)

    # 读取上次更新状态
    status = {
        'ipv4': {
            'last_ip': '',
            'last_update_domains': [],
        },
        'ipv6': {
            'last_ip': '',
            'last_update_domains': [],
        },
    }
    temp_dir = tempfile.gettempdir()
    status_file_path = os.path.join(temp_dir, 'cloudflare-ddns', 'status.dat')
    try:
        if os.path.exists(status_file_path):
            with open(status_file_path, 'r') as f:
                status = json.load(f)
    except Exception as e:
        logging.error("读取临时文件status.dat错误: %s", str(e))

    updated_domains_ipv4 = None
    updated_domains_ipv6 = None
    if settings['ipv4_ddns']:
        if ipv4:
            updated_domains_ipv4 = process_update(status['ipv4'], 'A', ipv4)
        else:
            logging.error("无法获取网络接口[%s]IPv4地址", settings['network_interface'])
    if settings['ipv6_ddns']:
        if ipv6:
            updated_domains_ipv6 = process_update(status['ipv6'], 'AAAA', ipv6)
        else:
            logging.error("无法获取网络接口[%s]IPv6地址", settings['network_interface'])

    status = {
        'ipv4': {
            'last_ip': ipv4,
            'last_update_domains': updated_domains_ipv4,
        },
        'ipv6': {
            'last_ip': ipv6,
            'last_update_domains': updated_domains_ipv6,
        },
    }

    # 回写更新记录
    try:
        os.makedirs(os.path.dirname(status_file_path), exist_ok=True)
        with open(status_file_path, 'w') as fp:
            json.dump(status, fp)
    except Exception as e:
        logging.error("写入临时文件status.dat错误: %s", str(e))


def get_network_addresses(interface):
    ipv4 = None
    ipv6 = None
    interfaces = psutil.net_if_addrs()
    if interface in interfaces:
        ipv4s = []
        ipv6s = []
        for addr in interfaces[interface]:
            if addr.family == socket.AF_INET:
                ipv4s.append(addr.address)
            elif addr.family == socket.AF_INET6:
                ipv6s.append(addr.address.split('%')[0])
        ipv4 = get_public_ipv4(ipv4s)
        ipv6 = get_public_ipv6(ipv6s)
    return ipv4, ipv6


def get_public_ipv4(addrs):
    for addr in addrs:
        ip_obj = ipaddress.IPv4Address(addr)
        if ip_obj.is_global:
            return addr
    return None


def get_public_ipv6(addrs):
    for addr in addrs:
        ip_obj = ipaddress.IPv6Address(addr)
        if ip_obj.is_global:
            return addr
    return None


def process_update(status, record_type, ip):
    # 判断哪些域名需要更新
    updated_domains = status.get('last_update_domains')
    to_update_domains = []
    if status.get('last_ip') == ip:
        for domain in settings['domain_names']:
            if domain not in updated_domains:
                to_update_domains.append(domain)
    else:
        to_update_domains = settings['domain_names']
    if not to_update_domains:
        logging.info("没有需要更新的域名")
    else:
        # 执行域名更新
        for domain in to_update_domains:
            try:
                dns_record_api_rsp = get_dns_record(record_type, domain, settings['zone_id'],
                                                    settings['authorization_key'])
                dns_record_id = dns_record_api_rsp[0].get('id') if len(dns_record_api_rsp) > 0 else None
                if not dns_record_id:
                    # 域名记录不存在，创建新的DNS记录
                    logging.info("域名[%s]不存在，创建新的DNS记录 -> [%s]", domain, ip)
                    create_dns_record(record_type, domain, ip, settings['zone_id'], settings['authorization_key'])
                else:
                    # 更新DNS记录
                    logging.info("域名[%s]存在，更新DNS记录 -> [%s]", domain, ip)
                    update_dns_record(record_type, domain, ip, settings['zone_id'], dns_record_id,
                                      settings['authorization_key'])
                updated_domains.append(domain)
            except Exception as e:
                logging.error("更新DNS记录异常: %s", str(e))
    return updated_domains


def config_logging():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s  %(filename)s : %(levelname)s  %(message)s',
        datefmt='%Y-%m-%d %A %H:%M:%S',
    )


def check_settings():
    if not settings['zone_id']:
        logging.critical('配置校验不通过: zone_id不能为空')
        exit(1)
    if not settings['authorization_key']:
        logging.critical('配置校验不通过: authorization_key不能为空')
        exit(1)
    if not settings['network_interface']:
        logging.critical('配置校验不通过: network_interface不能为空')
        exit(1)
    if not settings['domain_names']:
        logging.critical('配置校验不通过: domain_names需要配置一个以上的域名')
        exit(1)


if __name__ == '__main__':
    # 配置日志模块
    config_logging()
    # 检查配置
    check_settings()
    # 启动定时任务
    scheduler = BlockingScheduler()
    scheduler.add_job(main, 'interval', seconds=settings['update_interval_seconds'])
    scheduler.start()
