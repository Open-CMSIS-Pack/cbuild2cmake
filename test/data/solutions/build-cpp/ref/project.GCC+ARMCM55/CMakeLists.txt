cmake_minimum_required(VERSION 3.22)

set(CONTEXT project.GCC+ARMCM55)
set(TARGET ${CONTEXT})
set(OUT_DIR "${SOLUTION_ROOT}/out/project/ARMCM55/GCC")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(LD_SCRIPT "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/ARMCM55_gcc.ld")
set(LD_SCRIPT_PP ${LD_SCRIPT})

# Processor Options
set(CPU Cortex-M55)
set(FPU DP_FPU)
set(DSP DSP)
set(MVE FP_FVE)

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES C CXX)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}")
set(CMAKE_C_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_C_IMPLICIT_INCLUDE_DIRECTORIES})
set(CMAKE_CXX_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_CXX_IMPLICIT_INCLUDE_DIRECTORIES})

# Setup context
add_executable(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX ".elf" OUTPUT_NAME "project")
set_target_properties(${CONTEXT} PROPERTIES RUNTIME_OUTPUT_DIRECTORY ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes
add_library(${CONTEXT}_INCLUDES INTERFACE)
target_include_directories(${CONTEXT}_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55"
  "${SOLUTION_ROOT}/project/RTE/_GCC_ARMCM55"
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include"
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM55/Include"
)

# Defines
add_library(${CONTEXT}_DEFINES INTERFACE)
target_compile_definitions(${CONTEXT}_DEFINES INTERFACE
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ARMCM55
    _RTE_
  >
)

# Compile options
target_compile_options(${CONTEXT}_GLOBAL INTERFACE
  $<$<COMPILE_LANGUAGE:C>:
    SHELL:${CC_CPU}
    SHELL:${CC_FLAGS}
    -std=gnu11
    -masm-syntax-unified
    -fomit-frame-pointer
    -ffunction-sections
    -fdata-sections
  >
  $<$<COMPILE_LANGUAGE:CXX>:
    SHELL:${CXX_CPU}
    SHELL:${CXX_FLAGS}
    -masm-syntax-unified
    -fomit-frame-pointer
    -ffunction-sections
    -fdata-sections
  >
)

# Add groups and components
include("groups.cmake")
include("components.cmake")

target_link_libraries(${CONTEXT} PUBLIC
  ${CONTEXT}_GLOBAL
  Group_Source
  ARM_CMSIS_CORE_6_0_0
  ARM_Device_Startup_C_Startup_2_2_0
)
get_target_property(LINK_LIBRARIES ${CONTEXT} LINK_LIBRARIES)
set_target_properties(${CONTEXT} PROPERTIES LINK_LIBRARIES "-Wl,--start-group;${LINK_LIBRARIES};-Wl,--end-group")

# Linker options
target_link_options(${CONTEXT} PUBLIC
  SHELL:${LD_CPU}
  SHELL:${_LS}"${LD_SCRIPT_PP}"
  --specs=nano.specs
  --specs=rdimon.specs
  -Wl,-Map=${SOLUTION_ROOT}/out/project/ARMCM55/GCC/project.elf.map
  -Wl,--gc-sections
)
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})
