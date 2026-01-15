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

# Compilation database
add_custom_target(database
  COMMAND cmake -E env ${ENV_VARS} west build -b ${WEST_BOARD} -d "${OUT_DIR}" -p auto --cmake-only "${WEST_APP}" -- ${WEST_DEFS}
  USES_TERMINAL
)

# West build
add_custom_target(west
  COMMAND cmake -E env ${ENV_VARS} west build -b ${WEST_BOARD} -d "${OUT_DIR}" -p auto "${WEST_APP}" -- ${WEST_DEFS}
  USES_TERMINAL
)
