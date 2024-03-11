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
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/startup_ARMCM0.c"
  "${SOLUTION_ROOT}/project/RTE/Device/ARMCM0/system_ARMCM0.c"
)
target_link_libraries(ARM_Device_Startup_C_Startup_2_2_0 PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)

# component ARM::TestClass:TestGlobal@1.0.0
add_library(ARM_TestClass_TestGlobal_1_0_0 OBJECT
  "${SOLUTION_ROOT}/pack/Files/test1.c"
)
target_link_libraries(ARM_TestClass_TestGlobal_1_0_0 PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)

# component ARM::TestClass:TestLocal@1.0.0
add_library(ARM_TestClass_TestLocal_1_0_0 OBJECT
  "${SOLUTION_ROOT}/pack/Files/test2.c"
)
target_compile_options(ARM_TestClass_TestLocal_1_0_0 PRIVATE
  SHELL:${_PI}"${SOLUTION_ROOT}/pack/Files/header2.h"
  SHELL:${_PI}"${SOLUTION_ROOT}/project/RTE/TestClass/config-header2.h"
  SHELL:${_PI}"${SOLUTION_ROOT}/project/RTE/_GCC_ARMCM0/Pre_Include_TestClass_TestLocal.h"
)
target_link_libraries(ARM_TestClass_TestLocal_1_0_0 PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)
