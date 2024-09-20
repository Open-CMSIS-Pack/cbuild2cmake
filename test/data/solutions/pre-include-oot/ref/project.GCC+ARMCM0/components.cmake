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

# component ARM::TestClass:TestGlobal@1.0.0
add_library(ARM_TestClass_TestGlobal_1_0_0 OBJECT
  "${SOLUTION_ROOT}/pack/Files/test1.c"
)
target_include_directories(ARM_TestClass_TestGlobal_1_0_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(ARM_TestClass_TestGlobal_1_0_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(ARM_TestClass_TestGlobal_1_0_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)

# component ARM::TestClass:TestLocal@1.0.0
add_library(ARM_TestClass_TestLocal_1_0_0 OBJECT
  "${SOLUTION_ROOT}/pack/Files/test2.c"
)
target_include_directories(ARM_TestClass_TestLocal_1_0_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(ARM_TestClass_TestLocal_1_0_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(ARM_TestClass_TestLocal_1_0_0 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
  $<$<COMPILE_LANGUAGE:C,CXX>:
    "SHELL:${_PI}\"${SOLUTION_ROOT}/pack/Files/header2.h\""
    "SHELL:${_PI}\"${SOLUTION_ROOT}/project/RTE/TestClass/config-header2.h\""
    "SHELL:${_PI}\"${SOLUTION_ROOT}/project/RTE/_GCC_ARMCM0/Pre_Include_TestClass_TestLocal.h\""
  >
)
