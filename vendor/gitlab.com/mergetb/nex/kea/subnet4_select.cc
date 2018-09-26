/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

#include <hooks/hooks.h>

#include "nex.h"

using isc::hooks::CalloutHandle;

using namespace nex;

extern "C" {

  int subnet4_select(CalloutHandle &handle) {

    LOG_DEBUG(nex_logger, NEX_DEBUG_TRACE, LOG_NEX).arg("subnet4_select: enter");

    LOG_DEBUG(nex_logger, NEX_DEBUG_TRACE, LOG_NEX).arg("subnet4_select: exit");
    return 0;

  }

}
