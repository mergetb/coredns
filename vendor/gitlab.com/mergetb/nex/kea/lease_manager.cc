/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

#include <vector>
#include <string>
#include <bitset>
#include <stdexcept>
#include <sstream>
#include <memory>
#include <map>

#include <boost/algorithm/string/classification.hpp>

#include <etcd/Client.hpp>
#include <hooks/hooks.h>
#include <dhcp/pkt4.h>
#include <dhcp/option_string.h>
#include <dhcp/option_int.h>
#include <dhcp/option_int_array.h>
#include <asiolink/io_error.h>
#include <asiolink/io_address.h>

#include "json.hpp"
#include "lease_manager.h"
#include "nex.h"

using std::string;
using std::vector;
using std::bitset;
using std::exception;
using std::runtime_error;
using std::to_string;
using std::stringstream;
using std::make_shared;
using std::hex;
using std::unique_ptr;
using std::map;

using isc::hooks::CalloutHandle;
using isc::dhcp::Pkt4Ptr;
using isc::dhcp::HWAddrPtr;
using isc::dhcp::Option;
using isc::dhcp::OptionUint32;
using isc::dhcp::OptionUint8Array;
using isc::dhcp::OptionBuffer;
using isc::dhcp::OptionPtr;
using isc::dhcp::OptionString;
using isc::asiolink::IOAddress;
using isc::asiolink::IOError;
using isc::dhcp::DHCPOptionType;

using json = nlohmann::json;

using namespace nex;

string LeaseManager::getStaticIp4(string mac) {

  try {
    /* see if there is a static ip associated with the mac */
    string k = "/ifx/"+mac+"/ip4";
    etcd::Response r = etcd.get(k).get();
    if(r.error_code()) {
      throw runtime_error{"static ip not found"};
    }
    string ip4 = r.value().as_string();

    /* get the associated network info and set options */
    k = "/ifx/"+mac+"/net";
    r = etcd.get(k).get();
    if(r.error_code()) {
      throw runtime_error{"network name not found"};
    }
    Network net{r.value().as_string()};
    setOpts(net);

    NEX_INFO("getStaticIp4 assign: " + mac+"::"+ip4);
    return ip4;
  }
  catch(exception &e) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "getStaticIp4: "+string(e.what()));
    return "";
  }

}

// TODO: If we were to go the route of implementing a kea-etcd backend much of
// this would be handled transparently within kea.
string LeaseManager::getDynamicIp4(string mac) {

  /* get a list of all the dynamic networks */
  vector<string> names = getNets();

  /* iterate through the mac ranges associated with each dynamic net */
  for(auto name : names) {
    try {
      Network net{name};
      if(macBelongsToNet(mac, net)) {
        return leaseNetAddress(net, mac);
      }
    }
    catch(exception &e) {
      NEX_DEBUG(NEX_DEBUG_BASIC, "dynamic key search: "+string(e.what()));
      continue;
    }

  }
  return "";
}


size_t LeaseManager::poolIndex(string key) {
  size_t pos = key.rfind("/");
  string idx_str = key.substr(pos+1);
  return stoul(idx_str, nullptr, 10);
}

void LeaseManager::setOpts(Network &net) {

  vector<OptionPtr> opts{};

  /* standard options */

  // subnet mask
  uint32_t mask = net.subnet.mask();
  auto subnet_op = 
    boost::make_shared<OptionUint32>(
        Option::Universe::V4,
        DHCPOptionType::DHO_SUBNET_MASK, 
        mask);
  opts.push_back(subnet_op);
  NEX_DEBUG(NEX_DEBUG_BASIC, "subnet mask: " + to_string(mask));

  // gateway
  if(net.gateways) {
    vector<uint8_t> buf;
    for(auto g: *net.gateways) {
      try {
        IOAddress a{g};
        vector<uint8_t> abuf = a.toBytes();
        buf.insert(buf.end(), abuf.begin(), abuf.end());
      }
      catch(exception &e) {
        NEX_DEBUG(NEX_DEBUG_BASIC, "bad gateway: " + g);
        continue;
      }
    }
    opts.push_back(
        boost::make_shared<OptionUint8Array>(
          Option::Universe::V4, 
          DHCPOptionType::DHO_ROUTERS, 
          buf));
  }

  // name server
  if(net.nameservers) {
    vector<uint8_t> buf;
    for(auto n : *net.nameservers) {
      try {
        IOAddress a{n};
        vector<uint8_t> abuf = a.toBytes();
        buf.insert(buf.end(), abuf.begin(), abuf.end());
      }
      catch(exception &e) {
        NEX_DEBUG(NEX_DEBUG_BASIC, "bad nameserver: " + n);
        continue;
      }
    }
    opts.push_back(
        boost::make_shared<OptionUint8Array>(
          Option::Universe::V4, 
          DHCPOptionType::DHO_DOMAIN_NAME_SERVERS, 
          buf));
  }

  /* user defined options */
  
  if(net.options) {
    for(auto o : *net.options) {
      auto op = boost::make_shared<OptionString>(
          Option::Universe::V4,
          static_cast<uint16_t>(o.number),
          o.value
          );

      opts.push_back(op);
    }
  }

  handle->setContext("opts", opts);
}

/* get a list of all the dynamic networks */
vector<string> LeaseManager::getNets() 
{
  string k = "/nets";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_WARN("getNets: failed to get net list - "+r.error_message());
    return vector<string>{};
  }
  json j = json::parse(r.value().as_string());
  return j; //implicit conversion from json to vector
}

string LeaseManager::leaseNetAddress(Network &net, string mac)
{
  NEX_INFO("assigning "+mac+" to dynamic net "+net.name);

  /* check to make sure the network object has an ip range to assign from */
  if(!net.ip_range) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "network "+net.name+" has no ip range");
    return "";
  }
  IpRange iprange = *net.ip_range;

  /* fetch pool from db */
  etcd::Response r = getNetPool(net.name);

  /* iterate through the pool until the first unused address is found */
  uint32_t offset = r.keys().size();
  for(size_t i=0; i<r.keys().size(); i++) {
    size_t pi = poolIndex(r.keys()[i]);
    if(pi != i) {
      offset = pi;
      break;
    }
  }

  /* make sure there is not already a lease for this address */
  uint32_t lease_idx{0};
  for(size_t i=0; i<r.keys().size(); i++) {
    if(r.value(i).as_string() == mac) {
      lease_idx = i;
      break;
    }
  }
  /* lease already exists */
  if(lease_idx) {
    NEX_INFO("mac: "+mac+" already has lease "+to_string(lease_idx));
    IOAddress ip4_selected_addr{iprange.begin.toUint32() + lease_idx};
    setOpts(net);
    return ip4_selected_addr.toText();
  }

  /* calculate the target ip address */
  IOAddress ip4_selected_addr{iprange.begin.toUint32() + offset};
  if(iprange.end < ip4_selected_addr) {
    throw runtime_error{net.name+": ip range exhausted"};
  }
  string ip4_text = ip4_selected_addr.toText();
  NEX_INFO("dynamicLease: selected "+ip4_text+" for "+mac);

  /* create a etcd lease for the pool lease */
  setLease(net.name, offset, mac);
  setOpts(net);

  return ip4_selected_addr.toText();

}

etcd::Response LeaseManager::getNetPool(string net)
{
  string k = "/net/"+net+"/pool4/";         
  etcd::Response r = etcd.ls(k).get();         
  if(r.error_code()) {           
    throw runtime_error{"getNetPool: failed to get pool4 for "+net};
  }         
  return r;
}

void LeaseManager::setLease(string net, uint32_t offset, string mac)
{
  etcd::Response lease = etcd.leasegrant(3600).get();
  if(!lease.is_ok()) {
    throw runtime_error{"failed to get lease: "+lease.error_message()};
  }
  string k = "/net/"+net+"/pool4/"+to_string(offset);
  NEX_DEBUG(NEX_DEBUG_TRACE, "setLease: lease "+k+": "+mac);
  etcd::Response r = etcd.set(k, mac, lease.value().lease()).get();
  if(!r.is_ok()) {
    throw runtime_error{"failed to create lease: "+r.error_message()};
  }
  stringstream ss;
  ss << hex << lease.value().lease();
  string lease_hex = ss.str();
  NEX_INFO("pkt4_lease: " +lease_hex);
}

bool LeaseManager::macBelongsToNet(string mac, Network &net)
{
  // check if mac is within range
  if(net.mac_range) {
    MacRange mr = *net.mac_range;
    unsigned long maci = mac2Int(mac);
    if (maci > mr.begin && maci < mr.end) {
      return true;
    }
  }

  // check if mac is associated with a member
  if(net.members) {
    for(const auto &m : *net.members) {
      if(m.mac == mac) {
        NEX_DEBUG(NEX_DEBUG_TRACE, "membership "+net.name+": "+m.mac+" == "+mac);
        return true;
      }
    }
  } else {
    NEX_DEBUG(NEX_DEBUG_TRACE, net.name+" has no members");
  }

  return false;
}

LeaseManager::LeaseManager(string db, CalloutHandle &handle)
  : etcd{db},
    handle{&handle}
{
}

