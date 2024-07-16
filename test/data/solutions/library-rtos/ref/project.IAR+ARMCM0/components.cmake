# components.cmake

# component ARM::CMSIS:CORE@6.1.0
add_library(ARM_CMSIS_CORE_6_1_0 INTERFACE)
target_include_directories(ARM_CMSIS_CORE_6_1_0 INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
  ${CMSIS_PACK_ROOT}/ARM/CMSIS/6.1.0/CMSIS/Core/Include
)
target_compile_definitions(ARM_CMSIS_CORE_6_1_0 INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)

# component ARM::CMSIS:OS Tick:SysTick@1.0.5
add_library(ARM_CMSIS_OS_Tick_SysTick_1_0_5 OBJECT
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.1.0/CMSIS/RTOS2/Source/os_systick.c"
)
target_include_directories(ARM_CMSIS_OS_Tick_SysTick_1_0_5 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(ARM_CMSIS_OS_Tick_SysTick_1_0_5 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(ARM_CMSIS_OS_Tick_SysTick_1_0_5 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)

# component ARM::CMSIS:RTOS2:Keil RTX5&Library@5.9.0
add_library(ARM_CMSIS_RTOS2_Keil_RTX5_Library_5_9_0 OBJECT
  "${CMSIS_PACK_ROOT}/ARM/CMSIS-RTX/5.9.0/Source/rtx_lib.c"
  "${SOLUTION_ROOT}/project/RTE/CMSIS/RTX_Config.c"
)
target_include_directories(ARM_CMSIS_RTOS2_Keil_RTX5_Library_5_9_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
  ${SOLUTION_ROOT}/project/RTE/CMSIS
  ${CMSIS_PACK_ROOT}/ARM/CMSIS-RTX/5.9.0/Include
)
target_compile_definitions(ARM_CMSIS_RTOS2_Keil_RTX5_Library_5_9_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(ARM_CMSIS_RTOS2_Keil_RTX5_Library_5_9_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
target_link_libraries(ARM_CMSIS_RTOS2_Keil_RTX5_Library_5_9_0 PUBLIC
  ${CMSIS_PACK_ROOT}/ARM/CMSIS-RTX/5.9.0/Library/IAR/RTX_V6M.a
)

# component ARM::Device:Startup&C Startup@2.2.0
add_library(ARM_Device_Startup_C_Startup_2_2_0 OBJECT
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/startup_ARMCM0.c"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/system_ARMCM0.c"
)
target_include_directories(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
  ${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.1.0/Device/ARMCM0/Include
)
target_compile_definitions(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(ARM_Device_Startup_C_Startup_2_2_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
