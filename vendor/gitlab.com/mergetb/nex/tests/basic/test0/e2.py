import time, traceback

from avocado import Test
from avocado.utils import process
from netifaces import AF_INET
import netifaces as ni
import dns
import dns.resolver

class DhcpTestE2(Test):

    def test(s):
        try:
            process.system('sudo ip link set addr 00:00:77:44:AA:99 dev eth1')
            process.system('sudo ip link set up eth1')
            process.system('sudo dhclient -r eth1')
            process.system('sudo dhclient -1 eth1')
            info = ni.ifaddresses('eth1')[AF_INET][0]
            s.assertEqual(info['addr'], '10.2.0.13')
            s.assertEqual(info['netmask'], '255.255.0.0')

            gws = [x[0] for x in ni.gateways()[AF_INET]]
            s.assertTrue('10.2.0.1' in gws)
            resolvers = dns.resolver.get_default_resolver().nameservers
            s.assertTrue('10.2.0.1' in resolvers)
        except KeyError:
            s.fail("Address info for eth1 not found")
