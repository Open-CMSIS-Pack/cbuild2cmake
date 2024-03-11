# groups.cmake

# group Source1
add_library(Group_Source1 OBJECT
  "${SOLUTION_ROOT}/project/source1.c"
)
add_library(Group_Source1_INCLUDES INTERFACE)
target_include_directories(Group_Source1_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/inc1"
)
target_link_libraries(Group_Source1_INCLUDES INTERFACE
  ${CONTEXT}_INCLUDES
)
add_library(Group_Source1_DEFINES INTERFACE)
target_compile_definitions(Group_Source1_DEFINES INTERFACE
  DEF1=1
)
target_link_libraries(Group_Source1_DEFINES INTERFACE
  ${CONTEXT}_DEFINES
)
target_link_libraries(Group_Source1 PUBLIC
  ${CONTEXT}_GLOBAL
  Group_Source1_INCLUDES
  Group_Source1_DEFINES
)

# group Source2
add_library(Group_Source1_Source2 OBJECT
  "${SOLUTION_ROOT}/project/source2.c"
)
add_library(Group_Source1_Source2_INCLUDES INTERFACE)
target_include_directories(Group_Source1_Source2_INCLUDES INTERFACE
  "${SOLUTION_ROOT}/project/RTE/_IAR_ARMCM0"
  "${CMSIS_PACK_ROOT}/ARM/CMSIS/6.0.0/CMSIS/Core/Include"
  "${CMSIS_PACK_ROOT}/ARM/Cortex_DFP/1.0.0/Device/ARMCM0/Include"
  "${SOLUTION_ROOT}/project/inc2"
)
add_library(Group_Source1_Source2_DEFINES INTERFACE)
target_compile_definitions(Group_Source1_Source2_DEFINES INTERFACE
  ARMCM0
  _RTE_
  DEF2=1
)
target_link_libraries(Group_Source1_Source2 PRIVATE
  ${CONTEXT}_GLOBAL
  Group_Source1_Source2_INCLUDES
  Group_Source1_Source2_DEFINES
)

# group Main
add_library(Group_Main OBJECT
  "${SOLUTION_ROOT}/project/main.c"
)
target_link_libraries(Group_Main PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)
set_source_files_properties("${SOLUTION_ROOT}/project/main.c" PROPERTIES
  INCLUDE_DIRECTORIES "${SOLUTION_ROOT}/project/inc2"
  COMPILE_DEFINITIONS "DEF2"
)

