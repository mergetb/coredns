/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * This code is part of the Merge Testbed Framework
 * Copyright isi.edu 2018
 * Apache 2.0 License
 *
 * Authors:
 *  Ryan Goodfellow <rgoodfel@isi.edu>
 *~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

#include <log/logger_support.h>
#include <hooks/hooks.h>

using isc::hooks::LibraryHandle;
using isc::log::initLogger;

isc::log::Logger nex_logger("nex");

extern "C" {

  int load(LibraryHandle&) {
    initLogger(
        "nex",              /* logger name */
        isc::log::DEBUG,    /* severity level */
        99,                 /* debug level */
        NULL,               /* no local messages file */
        false               /* no buffering of messages */
    );
    return 0;
  }

  int unload() {
    return 0;
  }

}
