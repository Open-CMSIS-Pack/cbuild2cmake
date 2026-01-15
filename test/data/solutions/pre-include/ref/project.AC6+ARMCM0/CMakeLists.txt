cmake_minimum_required(VERSION 3.27)

# Roots
include("../roots.cmake")

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
project(${CONTEXT} LANGUAGES C)

# Enable color diagnostics
set(CMAKE_COLOR_DIAGNOSTICS ON)

# Compilation database
add_custom_target(database
  COMMAND ${CMAKE_COMMAND} -E make_directory "${OUT_DIR}"
  COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}"
)
set(CMAKE_C_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_C_IMPLICIT_INCLUDE_DIRECTORIES})

# Setup context
add_executable(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX ".axf" OUTPUT_NAME "project")
set_target_properties(${CONTEXT} PROPERTIES RUNTIME_OUTPUT_DIRECTORY ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes
target_include_directories(${CONTEXT} PUBLIC
  "${SOLUTION_ROOT}/project/RTE/_AC6_ARMCM0"
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include"
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM0/Include"
)

# Defines
target_compile_definitions(${CONTEXT} PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ARMCM0
    _RTE_
  >
)

# Compile options
target_compile_options(${CONTEXT} PUBLIC
  $<$<COMPILE_LANGUAGE:C>:
    "SHELL:${CC_CPU}"
    "SHELL:${CC_FLAGS}"
    "SHELL:-std=gnu11"
    "SHELL:-Wno-macro-redefined"
    "SHELL:-Wno-pragma-pack"
    "SHELL:-Wno-parentheses-equality"
    "SHELL:-Wno-license-management"
  >
  $<$<COMPILE_LANGUAGE:C,CXX>:
    "SHELL:${_PI}\"${SOLUTION_ROOT}/pack/Files/header1.h\""
    "SHELL:${_PI}\"${SOLUTION_ROOT}/project/RTE/TestClass/config-header1.h\""
    "SHELL:${_PI}\"${SOLUTION_ROOT}/project/RTE/_AC6_ARMCM0/Pre_Include_Global.h\""
  >
)

# Add groups and components
include("groups.cmake")
include("components.cmake")

target_link_libraries(${CONTEXT} PUBLIC
  Group_Source
  ARM_CMSIS_CORE_6_0_0
  ARM_Device_Startup_C_Startup_2_2_0
  ARM_TestClass_TestGlobal_1_0_0
  ARM_TestClass_TestLocal_1_0_0
)

# Linker options
target_link_options(${CONTEXT} PUBLIC
  "SHELL:${LD_CPU}"
  "SHELL:${_LS}\"${LD_SCRIPT_PP}\""
  "SHELL:--entry=Reset_Handler"
  "SHELL:--map"
  "SHELL:--info summarysizes"
  "SHELL:--summary_stderr"
  "SHELL:--diag_suppress=L6314W"
)
set(LD_DEPS ${LD_SCRIPT})
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS "${LD_DEPS}")
