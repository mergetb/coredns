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

  int lease4_select(CalloutHandle &handle) {

    LOG_DEBUG(nex_logger, NEX_DEBUG_TRACE, LOG_NEX).arg("lease4_select: enter");

    //handle.setStatus(CalloutHandle::NEXT_STEP_SKIP);

    LOG_DEBUG(nex_logger, NEX_DEBUG_TRACE, LOG_NEX).arg("lease4_select: exit");
    return 0;

  }

}
