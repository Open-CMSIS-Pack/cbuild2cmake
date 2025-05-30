cmake_minimum_required(VERSION 3.27)

# Roots
include("../roots.cmake")

set(CONTEXT project.IAR+ARMCM55)
set(TARGET ${CONTEXT})
set(OUT_DIR "${SOLUTION_ROOT}/out/project/ARMCM55/IAR")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(LD_SCRIPT "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/iar_linker_script.icf.src")
set(LD_REGIONS "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/regions_ARMCM55.h")
set(LD_SCRIPT_PP "${CMAKE_CURRENT_BINARY_DIR}/iar_linker_script.icf")

# Processor Options
set(CPU Cortex-M55)
set(FPU DP_FPU)
set(DSP DSP)
set(MVE FP_FVE)

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES CXX C)

# Compilation database
add_custom_target(database
  COMMAND ${CMAKE_COMMAND} -E make_directory "${OUT_DIR}"
  COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}"
)
set(CMAKE_CXX_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_CXX_IMPLICIT_INCLUDE_DIRECTORIES})
set(CMAKE_C_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_C_IMPLICIT_INCLUDE_DIRECTORIES})

# Setup context
add_executable(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX ".out" OUTPUT_NAME "project")
set_target_properties(${CONTEXT} PROPERTIES RUNTIME_OUTPUT_DIRECTORY ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes
target_include_directories(${CONTEXT} PUBLIC
  "${SOLUTION_ROOT}/project/RTE/_IAR_ARMCM55"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55"
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include"
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM55/Include"
)

# Defines
target_compile_definitions(${CONTEXT} PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ARMCM55
    _RTE_
  >
)

# Compile options
target_compile_options(${CONTEXT} PUBLIC
  $<$<COMPILE_LANGUAGE:C>:
    "SHELL:${CC_CPU}"
    "SHELL:${CC_FLAGS}"
    "SHELL:--dlib_config DLib_Config_Full.h"
  >
  $<$<COMPILE_LANGUAGE:CXX>:
    "SHELL:${CXX_CPU}"
    "SHELL:${CXX_FLAGS}"
    "SHELL:--dlib_config DLib_Config_Full.h"
  >
)

# Add groups and components
include("groups.cmake")
include("components.cmake")

target_link_libraries(${CONTEXT} PUBLIC
  Group_Source
  ARM_CMSIS_CORE_6_0_0
  ARM_Device_Startup_C_Startup_2_2_0
)

# Linker options
target_link_options(${CONTEXT} PUBLIC
  "SHELL:${LD_CPU}"
  "SHELL:${_LS}\"${LD_SCRIPT_PP}\""
  "SHELL:--semihosting"
  "SHELL:--map=${SOLUTION_ROOT}/out/project/ARMCM55/IAR/project.out.map"
)
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})

# Linker script pre-processing
add_custom_command(TARGET ${CONTEXT} PRE_LINK COMMAND ${CPP} ARGS ${CPP_ARGS_LD_SCRIPT} BYPRODUCTS ${LD_SCRIPT_PP})
