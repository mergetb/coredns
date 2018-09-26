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
#include <log/message_initializer.h>
#include <log/macros.h>
#include <messages.h>

#define NEX_DEBUG_TRACE 99
#define NEX_DEBUG_STD   47
#define NEX_DEBUG_BASIC 0

namespace nex {
  extern isc::log::Logger nex_logger;

  struct HostInfo {
    std::string name;
    std::string mac;
  };

  static inline void NEX_DEBUG(int level, std::string message) {
    LOG_DEBUG(nex_logger, NEX_DEBUG_TRACE, LOG_NEX).arg(message);
  }

  static inline void NEX_INFO(std::string message) {
    LOG_INFO(nex_logger, LOG_NEX).arg(message);
  }

  static inline void NEX_WARN(std::string message) {
    LOG_WARN(nex_logger, LOG_NEX).arg(message);
  }

}
