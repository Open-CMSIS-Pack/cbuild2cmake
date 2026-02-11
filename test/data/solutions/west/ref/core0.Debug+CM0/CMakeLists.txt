cmake_minimum_required(VERSION 3.27)

# Roots
include("../roots.cmake")

set(CONTEXT core0.Debug+CM0)
set(TARGET ${CONTEXT})
set(OUT_DIR "${SOLUTION_ROOT}/out/core0/CM0/Debug")
set(WEST_BOARD "rtetest dummy board")
set(WEST_APP "${SOLUTION_ROOT}/west/core0")

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES NONE)

set(WEST_DEFS
  -DCORE=0
  -DCONFIG_BUILD_OUTPUT_HEX=y
  -DTARGET-DEF=on
  -DCONFIG-DEBUG=y
)

# Enable color diagnostics
set(CMAKE_COLOR_DIAGNOSTICS ON)

# Environment variables
set(ZEPHYR_TOOLCHAIN_PATH "${REGISTERED_TOOLCHAIN_ROOT}/..")
cmake_path(ABSOLUTE_PATH ZEPHYR_TOOLCHAIN_PATH NORMALIZE OUTPUT_VARIABLE ZEPHYR_TOOLCHAIN_PATH)
set(ENV_VARS
  ARMCLANG_TOOLCHAIN_PATH="${ZEPHYR_TOOLCHAIN_PATH}"
  ZEPHYR_TOOLCHAIN_VARIANT="armclang"
)

# West setup and compilation database
add_custom_target(database DEPENDS "${OUT_DIR}/CMakeCache.txt")
add_custom_command(
  OUTPUT "${OUT_DIR}/CMakeCache.txt"
  BYPRODUCTS "${OUT_DIR}/compile_commands.json"
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Setup west build directory"
  COMMAND cmake -E env ${ENV_VARS} west build -b ${WEST_BOARD} -d "${OUT_DIR}" -p auto --cmake-only "${WEST_APP}" -- ${WEST_DEFS}
  COMMENT "" 
  DEPENDS "${CMAKE_CURRENT_LIST_FILE}"
  USES_TERMINAL
)

# West build
add_custom_target(west
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Run west build"
  COMMAND west build -d "${OUT_DIR}"
  COMMENT ""
  DEPENDS database
  USES_TERMINAL
)
