# groups.cmake

# group Source
add_library(Group_Source OBJECT
  "${SOLUTION_ROOT}/project/main.c"
)
target_link_libraries(Group_Source PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)

# group GCC-CLANG
add_library(Group_GCC-CLANG OBJECT
  "${SOLUTION_ROOT}/project/GCC/PreProcessed.S"
  "${SOLUTION_ROOT}/project/GCC/NonPreProcessed.s"
)
target_link_libraries(Group_GCC-CLANG PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)
set_source_files_properties("${SOLUTION_ROOT}/project/GCC/PreProcessed.S" PROPERTIES
  COMPILE_DEFINITIONS PRE_PROCESSED_DEF
)
set_source_files_properties("${SOLUTION_ROOT}/project/GCC/NonPreProcessed.s" PROPERTIES
  COMPILE_OPTIONS -DPRE_PROCESSED_DEF
)
