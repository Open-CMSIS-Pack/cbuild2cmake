cmake_minimum_required(VERSION 3.27)

# Roots
include("../roots.cmake")

set(CONTEXT core0.Debug+CM0)
set(TARGET ${CONTEXT})
set(OUT_DIR "${SOLUTION_ROOT}/out/core0/CM0/Debug")
set(CMAKE_SOURCE_DIR_NATIVE "${SOLUTION_ROOT}/cmake/core0")
set(CMAKE_GENERATOR_NATIVE "Ninja Multi-Config")

project(${CONTEXT} LANGUAGES NONE)

set(CMAKE_CONFIGURE_OPTIONS
  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
  -DCMAKE_BUILD_TYPE=Debug
)

# Native CMake setup and compilation database
add_custom_target(database DEPENDS "${OUT_DIR}/CMakeCache.txt")
add_custom_command(
  OUTPUT "${OUT_DIR}/CMakeCache.txt"
  BYPRODUCTS "${OUT_DIR}/compile_commands.json"
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Setup native CMake build directory"
  COMMAND ${CMAKE_COMMAND} -S "${CMAKE_SOURCE_DIR_NATIVE}" -B "${OUT_DIR}" -G "${CMAKE_GENERATOR_NATIVE}" ${CMAKE_CONFIGURE_OPTIONS}
  COMMENT ""
  DEPENDS "${CMAKE_CURRENT_LIST_FILE}"
  USES_TERMINAL
)

# Native CMake build
add_custom_target(cmake
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Run native CMake build"
  COMMAND ${CMAKE_COMMAND} --build "${OUT_DIR}" --target "core0-target"
  COMMENT ""
  DEPENDS database
  USES_TERMINAL
)
