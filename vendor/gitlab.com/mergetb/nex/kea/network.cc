/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

#include <stdexcept>
#include <string>
#include <vector>

#include <boost/algorithm/string/classification.hpp>
#include <boost/range/algorithm/remove_if.hpp>

#include "json.hpp"
#include "nex.h"
#include "network.h"

using std::string;
using std::vector;
using std::exception;
using std::runtime_error;

using isc::asiolink::IOAddress;

using json = nlohmann::json;

using namespace nex;


Network::Network(string name, string connstr)
  : name{name},
    etcd{connstr}
{
  pull();
}

void Network::pull()
{
  getSubnet();

  // optional properties
  getDomain();
  getIpRange();
  getMacRange();
  getGateways();
  getNameservers();
  getOptions();
  getMembers();

}


void Network::getSubnet()
{
  string k = "/net/"+name+"/subnet4";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    throw runtime_error{"failed to fetch subnet from db"};
  }
  vector<string> subnet_parts;
  string s4 = r.value().as_string();
  boost::split(subnet_parts, s4, boost::is_any_of("/"));
  if(subnet_parts.size() < 2) {
    throw runtime_error{"invalid subnet"};
  }

  subnet = Subnet{
    subnet_parts[0],
    static_cast<uint8_t>(stoul(subnet_parts[1], nullptr, 10))
  };
}

void Network::getDomain()
{
  string k = "/net/"+name+"/domain";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "domain property not found");
    return;
  } 

  domain = r.value().as_string();
}

void Network::getIpRange()
{
  string k = "/net/"+name+"/ip4_range";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "iprange property not found");
    return;
  } 

  json j;
  try { j = json::parse(r.value().as_string()); }
  catch(...) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to parse iprange");
    return;
  }
  string ip4_begin = j["begin"], ip4_end = j["end"];

  if(ip4_begin == "" || ip4_end == "") {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to read iprange");
    return;
  }

  ip_range = IpRange{
    IOAddress{ip4_begin},
    IOAddress{ip4_end}
  };
}

void Network::getMacRange()
{
  string k = "/net/"+name+"/mac_range";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "mac_range property not found");
    return;
  }
  json j;
  try { j = json::parse(r.value().as_string()); }
  catch(...) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to parse mac_range");
    return;
  }
  string mac_begin = j["begin"], mac_end = j["end"];

  if(mac_begin == "" || mac_end == "") {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to read mac range");
    return;
  }

  mac_range = MacRange{
    mac2Int(mac_begin),
    mac2Int(mac_end)
  };
}

void Network::getGateways()
{
  string k = "/net/"+name+"/gateways";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "gateways property not found");
    return;
  }
  try { 
    vector<string> gs = json::parse(r.value().as_string()); 
    gateways = gs;
  }
  catch(...) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to parse gateways");
    return;
  }
}

void Network::getNameservers()
{
  string k = "/net/"+name+"/nameservers";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "nameservers property not found");
    return;
  }
  try { 
    vector<string> ns = json::parse(r.value().as_string()); 
    nameservers = ns;
  }
  catch(...) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to parse nameservers");
    return;
  }
}

void Network::getOptions()
{
  string k = "/net/"+name+"/opts";
  etcd::Response r = etcd.get(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "options property not found");
    return;
  }

  try{ 
    vector<json> js = json::parse(r.value().as_string()); 
    vector<Opt> opts{};

    for(auto j : js) {
      int number = j["number"];
      string value = j["value"];

      if(number == 0 || value == "") {
        continue;
      }

      opts.emplace_back(number, value);
    }

    options = opts;
  }
  catch(...) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to parse options");
    return;
  }
}

void Network::getMembers()
{
  string k = "/net/"+name+"/members";
  etcd::Response r = etcd.ls(k).get();
  if(r.error_code()) {
    NEX_DEBUG(NEX_DEBUG_BASIC, name+": members property not found");
    return;
  }

  try {
    vector<Member> ms{};

    uint32_t offset = r.values().size();
    for(size_t i=0; i<r.values().size(); i++) {
      string mac = r.values()[i].as_string();  
      //TODO snag other properties from /ifx prefix
      ms.emplace_back(mac, "", "");
    }

    members = ms;
  }
  catch(...) {
    NEX_DEBUG(NEX_DEBUG_BASIC, "failed to parse members");
    return;
  }

}

/* parse mac as unsigned hexidecimal number */
unsigned long nex::mac2Int(string mac) {

  mac.erase(boost::remove_if(mac, boost::is_any_of(":")), mac.end());
  mac = "0x"+mac;
  return stoul(mac, nullptr, 16);

}

uint32_t Subnet::mask()
{
  return ((1ul<<prefix)-1) << (32-prefix);
}


