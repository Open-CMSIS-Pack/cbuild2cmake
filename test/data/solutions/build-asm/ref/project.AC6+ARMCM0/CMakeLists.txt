cmake_minimum_required(VERSION 3.22)

set(CONTEXT project.AC6+ARMCM0)
set(TARGET ${CONTEXT})
set(OUT_DIR "${SOLUTION_ROOT}/out/project/ARMCM0/AC6")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(LD_SCRIPT "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/ARMCM0_ac6.sct")
set(LD_SCRIPT_PP ${LD_SCRIPT})

# Processor Options
set(CPU Cortex-M0)
set(FPU NO_FPU)

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES C ASM)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}")
set(CMAKE_C_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_C_IMPLICIT_INCLUDE_DIRECTORIES})

# Setup context
add_executable(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX ".axf" OUTPUT_NAME "project")
set_target_properties(${CONTEXT} PROPERTIES RUNTIME_OUTPUT_DIRECTORY ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes
add_library(${CONTEXT}_INCLUDES INTERFACE)
target_include_directories(${CONTEXT}_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/RTE/_AC6_ARMCM0"
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include"
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM0/Include"
)

# Defines
add_library(${CONTEXT}_DEFINES INTERFACE)
target_compile_definitions(${CONTEXT}_DEFINES INTERFACE
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ARMCM0
    _RTE_
  >
)

# Compile options
target_compile_options(${CONTEXT}_GLOBAL INTERFACE
  $<$<COMPILE_LANGUAGE:ASM>:
    SHELL:${ASM_CPU}
    SHELL:${ASM_FLAGS}
    -masm=auto
  >
  $<$<COMPILE_LANGUAGE:C>:
    SHELL:${CC_CPU}
    SHELL:${CC_FLAGS}
    -std=gnu11
    -Wno-macro-redefined
    -Wno-pragma-pack
    -Wno-parentheses-equality
    -Wno-license-management
  >
)

# Add groups and components
include("groups.cmake")
include("components.cmake")

target_link_libraries(${CONTEXT} PUBLIC
  ${CONTEXT}_GLOBAL
  Group_Source
  Group_AC6
  ARM_CMSIS_CORE_6_0_0
  ARM_Device_Startup_C_Startup_2_2_0
)

# Linker options
target_link_options(${CONTEXT} PUBLIC
  SHELL:${LD_CPU}
  SHELL:${_LS}"${LD_SCRIPT_PP}"
  --entry=Reset_Handler
  --map
  --info summarysizes
  --summary_stderr
  --diag_suppress=L6314W
)
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})
