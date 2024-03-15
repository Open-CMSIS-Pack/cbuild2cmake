cmake_minimum_required(VERSION 3.22)

set(CONTEXT project.CLANG+ARMCM0)
set(TARGET ${CONTEXT})
set(OUT_DIR "${SOLUTION_ROOT}/out/project/ARMCM0/CLANG")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(LD_SCRIPT "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/clang_linker_script.ld.src")
set(LD_REGIONS "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/regions_ARMCM0.h")
set(LD_SCRIPT_PP "${CMAKE_CURRENT_BINARY_DIR}/clang_linker_script.ld")

# Processor Options
set(CPU Cortex-M0)
set(FPU NO_FPU)

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES C ASM)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}")

# Setup context
add_executable(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX ".elf" OUTPUT_NAME "project")
set_target_properties(${CONTEXT} PROPERTIES RUNTIME_OUTPUT_DIRECTORY ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes
add_library(${CONTEXT}_INCLUDES INTERFACE)
target_include_directories(${CONTEXT}_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/RTE/_CLANG_ARMCM0"
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
  >
  $<$<COMPILE_LANGUAGE:C>:
    SHELL:${CC_CPU}
    SHELL:${CC_FLAGS}
    -std=gnu11
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
  Group_GCC-CLANG
  ARM_CMSIS_CORE_6_0_0
  ARM_Device_Startup_C_Startup_2_2_0
)

# Linker options
target_link_options(${CONTEXT} PUBLIC
  SHELL:${LD_CPU}
  SHELL:${_LS}"${LD_SCRIPT_PP}"
  -lcrt0-semihost
  -lsemihost
  -Wl,-Map=${SOLUTION_ROOT}/out/project/ARMCM0/CLANG/project.elf.map
  -Wl,--gc-sections
)
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})

# Linker script pre-processing
add_custom_command(TARGET ${CONTEXT} PRE_LINK COMMAND ${CPP} ARGS ${CPP_ARGS_LD_SCRIPT} BYPRODUCTS ${LD_SCRIPT_PP})
