/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
#pragma once

#include <etcd/Client.hpp>
#include <hooks/hooks.h>

#include "network.h"

namespace nex {

  struct LeaseManager {

    LeaseManager(std::string db, isc::hooks::CalloutHandle &handle);

    etcd::Client etcd;
    isc::hooks::CalloutHandle *handle;

    std::string getStaticIp4(std::string mac);
    std::string getDynamicIp4(std::string mac);
    size_t poolIndex(std::string key);

    void setOpts(Network &net);
    std::vector<std::string> getNets();
    etcd::Response getNetPool(std::string net);
    bool macBelongsToNet(std::string mac, Network &net);
    std::string leaseNetAddress(Network &net, std::string mac);
    void setLease(std::string net, uint32_t offset, std::string mac);

  };

}
