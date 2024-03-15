# components.cmake

# component ARM::CMSIS:CORE@6.0.0
add_library(ARM_CMSIS_CORE_6_0_0 INTERFACE)
add_library(ARM_CMSIS_CORE_6_0_0_INCLUDES INTERFACE)
target_include_directories(ARM_CMSIS_CORE_6_0_0_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/RTE/_CLANG_ARMCM0"
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include"
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM0/Include"
)
add_library(ARM_CMSIS_CORE_6_0_0_DEFINES INTERFACE)
target_compile_definitions(ARM_CMSIS_CORE_6_0_0_DEFINES INTERFACE
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ARMCM0
    _RTE_
  >
)
target_link_libraries(ARM_CMSIS_CORE_6_0_0 INTERFACE
  ${CONTEXT}_GLOBAL
  ARM_CMSIS_CORE_6_0_0_INCLUDES
  ARM_CMSIS_CORE_6_0_0_DEFINES
)

# component ARM::Device:Startup&C Startup@2.2.0
add_library(ARM_Device_Startup_C_Startup_2_2_0 OBJECT
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/startup_ARMCM0.c"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/system_ARMCM0.c"
)
add_library(ARM_Device_Startup_C_Startup_2_2_0_INCLUDES INTERFACE)
target_include_directories(ARM_Device_Startup_C_Startup_2_2_0_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/inc2"
)
add_library(ARM_Device_Startup_C_Startup_2_2_0_DEFINES INTERFACE)
target_compile_definitions(ARM_Device_Startup_C_Startup_2_2_0_DEFINES INTERFACE
  $<$<COMPILE_LANGUAGE:C,CXX>:
    DEF2=1
  >
)
target_link_libraries(ARM_Device_Startup_C_Startup_2_2_0 PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ARM_Device_Startup_C_Startup_2_2_0_INCLUDES
  ${CONTEXT}_DEFINES
  ARM_Device_Startup_C_Startup_2_2_0_DEFINES
)
