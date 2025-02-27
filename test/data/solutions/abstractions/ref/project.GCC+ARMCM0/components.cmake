# components.cmake

# component ARM::CMSIS:CORE@6.1.0
add_library(ARM_CMSIS_CORE_6_1_0 INTERFACE)
target_include_directories(ARM_CMSIS_CORE_6_1_0 INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.1.0/CMSIS/Core/Include"
)
target_compile_definitions(ARM_CMSIS_CORE_6_1_0 INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_link_libraries(ARM_CMSIS_CORE_6_1_0 INTERFACE
  ${CONTEXT}_ABSTRACTIONS
)

# component ARM::Device:Startup&C Startup@2.2.0
add_library(ARM_Device_Startup_C_Startup_2_2_0 OBJECT
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/startup_ARMCM0.c"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/system_ARMCM0.c"
)
target_include_directories(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.1.0/Device/ARMCM0/Include"
)
target_compile_definitions(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
add_library(ARM_Device_Startup_C_Startup_2_2_0_ABSTRACTIONS INTERFACE)
cbuild_set_options_flags(CC "none" "off" "" "" CC_OPTIONS_FLAGS_ARM_Device_Startup_C_Startup_2_2_0)
target_compile_options(ARM_Device_Startup_C_Startup_2_2_0_ABSTRACTIONS INTERFACE
  $<$<COMPILE_LANGUAGE:C>:
    "SHELL:${CC_OPTIONS_FLAGS_ARM_Device_Startup_C_Startup_2_2_0}"
  >
)
target_compile_options(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
target_link_libraries(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  ARM_Device_Startup_C_Startup_2_2_0_ABSTRACTIONS
)
