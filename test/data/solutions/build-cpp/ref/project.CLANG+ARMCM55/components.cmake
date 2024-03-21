# components.cmake

# component ARM::CMSIS:CORE@6.0.0
add_library(ARM_CMSIS_CORE_6_0_0 INTERFACE)
target_link_libraries(ARM_CMSIS_CORE_6_0_0 INTERFACE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)

# component ARM::Device:Startup&C Startup@2.2.0
add_library(ARM_Device_Startup_C_Startup_2_2_0 OBJECT
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/startup_ARMCM55.c"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55/system_ARMCM55.c"
)
target_include_directories(ARM_Device_Startup_C_Startup_2_2_0
  PUBLIC
    "${SOLUTION_ROOT}/project/RTE/Device/ARMCM55"
)
target_link_libraries(ARM_Device_Startup_C_Startup_2_2_0 PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)
