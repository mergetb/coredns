/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

#include <hooks/hooks.h>
#include <dhcp/pkt4.h>
#include <asiolink/io_address.h>

#include "json.hpp"
#include "lease_manager.h"
#include "nex.h"

using std::string;
using isc::hooks::CalloutHandle;
using isc::dhcp::Pkt4Ptr;
using isc::dhcp::HWAddrPtr;

using namespace nex;


extern "C" {

  int pkt4_receive(CalloutHandle &handle) {

    NEX_DEBUG(NEX_DEBUG_TRACE, "pkt4_receive: enter");

    LeaseManager lm{"http://db:2379", handle};

    /* fetch the mac from the query */
    Pkt4Ptr query4_ptr;
    handle.getArgument("query4", query4_ptr);
    HWAddrPtr hwaddr_ptr = query4_ptr->getHWAddr();
    string hwaddr = hwaddr_ptr->toText(false);
    NEX_DEBUG(NEX_DEBUG_BASIC, "pkt4_receive: MAC " + hwaddr);
    handle.setContext("hwaddr", hwaddr);

    /* first check if there is a static ip associated with this mac */
    string ip4 = lm.getStaticIp4(hwaddr);
    if(ip4 != "") {
      handle.setContext("ipaddr", ip4);
      return 0;
    }

    /* next check if there is a dynamic range this mac falls into */
    ip4 = lm.getDynamicIp4(hwaddr);
    if(ip4 != "") {
      handle.setContext("ipaddr", ip4);
      return 0;
    }

    /* is we are here - we know nothing about this mac */
    return 1;
  }

}

