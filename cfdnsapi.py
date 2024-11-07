import requests


def get_dns_record(record_type, domain, zone_id, authorization_key):
    """精确匹配IP和域名查询DNS记录信息"""
    url = 'https://api.cloudflare.com/client/v4/zones/%s/dns_records' % zone_id
    params = {
        'type': record_type,
        'name': domain,
        'match': 'all'
    }
    headers = {
        'Authorization': 'Bearer %s' % authorization_key,
        'Content-Type': 'application/json'
    }
    rsp = requests.get(url, params, headers=headers)
    rsp.encoding = "utf-8"
    rsp_json = rsp.json()
    if not rsp_json.get('success'):
        code, message = get_rsp_err_desc(rsp_json)
        raise Exception('API调用失败：错误码：%s 错误描述：%s' % (code, message))
    return rsp_json.get('result')


def create_dns_record(record_type, domain, ip, zone_id, authorization_key):
    """创建DNS记录"""
    url = 'https://api.cloudflare.com/client/v4/zones/%s/dns_records' % zone_id
    json_data = {
        'type': record_type,
        'name': domain,
        'content': ip,
        'ttl': 120
    }
    headers = {
        'Authorization': 'Bearer %s' % authorization_key,
        'Content-Type': 'application/json'
    }
    rsp = requests.post(url, json=json_data, headers=headers)
    rsp.encoding = "utf-8"
    rsp_json = rsp.json()
    if not rsp_json.get('success'):
        code, message = get_rsp_err_desc(rsp_json)
        raise Exception('API调用失败：错误码：%s 错误描述：%s' % (code, message))
    return rsp_json.get('result')


def update_dns_record(record_type, domain, ip, zone_id, dns_id, authorization_key):
    """更新DNS记录"""
    url = 'https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s' % (zone_id, dns_id)
    json_data = {
        'type': record_type,
        'name': domain,
        'content': ip,
        'ttl': 120
    }
    headers = {
        'Authorization': 'Bearer %s' % authorization_key,
        'Content-Type': 'application/json'
    }
    rsp = requests.put(url, json=json_data, headers=headers)
    rsp.encoding = "utf-8"
    rsp_json = rsp.json()
    if not rsp_json.get('success'):
        code, message = get_rsp_err_desc(rsp_json)
        raise Exception('API调用失败：错误码：%s 错误描述：%s' % (code, message))
    return rsp_json.get('result')


def get_rsp_err_desc(rsp):
    if rsp and rsp.get('errors'):
        error0 = rsp['errors'][0]
        code = error0.get('code', 'None')
        message = error0.get('message', 'None')
    else:
        code, message = 'None', 'None'
    return code, message
