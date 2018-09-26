
/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
#pragma once

#include <string>
#include <vector>
#include <optional>

#include <etcd/Client.hpp>
#include <asiolink/io_address.h>

namespace nex {

  struct MacRange {
    unsigned long begin, end;
  };

  struct IpRange {
    isc::asiolink::IOAddress begin, end;

    /* default range is from 0.0.0.0/0 ~ 0.0.0.0/0 */
    IpRange(): begin{0}, end{0} {}

    IpRange(isc::asiolink::IOAddress begin, isc::asiolink::IOAddress end)
      : begin{begin},
        end{end}
    {}
  };

  struct Subnet {
    std::string address;
    uint8_t prefix{0};

    Subnet() = default;
    Subnet(std::string address, uint8_t prefix)
      : address{address},
        prefix{prefix}
    {}

    uint32_t mask();
  };

  struct Opt {
    int number;
    std::string value;

    Opt(int number, std::string value) 
      : number{number}, 
        value{value} 
    {}
  };

  struct Member {
    std::string mac,
                ip4,
                name;

    Member(std::string mac, std::string ip4, std::string name)
      : mac{mac},
        ip4{ip4},
        name{name}
    {}
  };

  struct Network {
    /* required */
    std::string                                 name;
    Subnet                                      subnet;

    /* optional */
    std::optional<std::string>                  domain;
    std::optional<IpRange>                      ip_range;
    std::optional<MacRange>                     mac_range;
    std::optional<std::vector<std::string>>     gateways;
    std::optional<std::vector<std::string>>     nameservers;
    std::optional<std::vector<Opt>>             options;
    std::optional<std::vector<Member>>          members;

    /* throws runtime error if required network info cannot be fetched */
    Network(std::string name, std::string connstr="http://db:2379");

    private:
      etcd::Client etcd;

      void pull();
      void getSubnet();
      void getDomain();
      void getIpRange();
      void getMacRange();
      void getGateways();
      void getNameservers();
      void getOptions();
      void getMembers();
  };

  unsigned long mac2Int(std::string mac);

}
