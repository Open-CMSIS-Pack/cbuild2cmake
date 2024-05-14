# components.cmake

# component ARM::CMSIS:CORE@6.0.0
add_library(ARM_CMSIS_CORE_6_0_0 INTERFACE)
target_include_directories(ARM_CMSIS_CORE_6_0_0 INTERFACE
  ${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(ARM_CMSIS_CORE_6_0_0 INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)

# component ARM::Device:Startup&C Startup@2.2.0
add_library(ARM_Device_Startup_C_Startup_2_2_0 OBJECT
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/startup_ARMCM55.c"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/system_ARMCM55.c"
)
target_include_directories(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  ${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM55/Include
  ${SOLUTION_ROOT}/project/RTE/Device/ARMCM55
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
