/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

#include <string>
#include <vector>
#include <stdexcept>

#include <hooks/hooks.h>
#include <dhcp/pkt4.h>
#include <dhcp/option.h>
#include <asiolink/io_error.h>
#include <boost/algorithm/string.hpp>

#include "nex.h"

using std::string;
using std::vector;
using std::invalid_argument;

using isc::hooks::CalloutHandle;
using isc::hooks::NoSuchCalloutContext;
using isc::dhcp::Pkt4Ptr;
using isc::dhcp::OptionPtr;
using isc::asiolink::IOAddress;
using isc::asiolink::IOError;

using namespace nex;

extern "C" {

  int pkt4_send(CalloutHandle &handle) {

    NEX_DEBUG(NEX_DEBUG_TRACE, "pkt4_send: enter");

    string hwaddr;
    string ipaddr;
    try {
      handle.getContext("hwaddr", hwaddr);
      handle.getContext("ipaddr", ipaddr);

      Pkt4Ptr response4_ptr;
      handle.getArgument("response4", response4_ptr);

      /* get options - if any */
      vector<OptionPtr> opts{};
      try {
        handle.getContext("opts", opts);
      }
      catch(...) {
        // its ok not to have any options
      }

      for(const auto & opt : opts) {
        NEX_DEBUG(NEX_DEBUG_TRACE, 
            "pkt4_send: delivering option - "+opt->toText());

        /* need to delete first b/c add will throw if already present, and we
         * want to just clobber and move on anyhow */
        response4_ptr->delOption(opt->getType());
        response4_ptr->addOption(opt);
      }

      IOAddress yiaddr{ipaddr};
      response4_ptr->setYiaddr(yiaddr);

      NEX_DEBUG(NEX_DEBUG_BASIC, "pkt4_send: " + hwaddr+"::"+ipaddr);

    } 
    catch(const NoSuchCalloutContext&) 
    {
      // no hwaddr here
    } 
    catch(const IOError &e) 
    {
      NEX_WARN("pkt4_send: " + string(e.what()));
    }
    catch(invalid_argument &e)
    {
      NEX_WARN("pkt4_send: " + string(e.what()));
    }

    NEX_DEBUG(NEX_DEBUG_TRACE, "pkt4_send: exit");
    return 0;

  }

}
