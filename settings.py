import os

from dotenv import load_dotenv


def clean(value):
    return value.strip() if value and value.strip() else None


load_dotenv()
settings = {
    'log_level': clean(os.getenv('LOG_LEVEL')),
    'zone_id': clean(os.getenv('ZONE_ID')),
    'authorization_key': clean(os.getenv('AUTHORIZATION_KEY')),
    'network_interface': clean(os.getenv('NETWORK_INTERFACE')),
    'domain_names': [name.strip() for name in os.getenv('DOMAIN_NAMES', '').split(',') if name.strip()],
    'ipv4_ddns': True if clean(os.getenv('IPV4_DDNS')) == '1' else False,
    'ipv6_ddns': True if clean(os.getenv('IPV6_DDNS')) == '1' else False,
    'update_interval_seconds': int(os.getenv('UPDATE_INTERVAL_SECONDS', 120)),
}
