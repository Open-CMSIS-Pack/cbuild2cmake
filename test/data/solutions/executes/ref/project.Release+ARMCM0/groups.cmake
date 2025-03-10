# groups.cmake

# group Source
add_library(Group_Source OBJECT
  "${SOLUTION_ROOT}/project/source0.c"
  "${SOLUTION_ROOT}/project/source1.c"
)
target_include_directories(Group_Source PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Source PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(Group_Source PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
set_source_files_properties("${SOLUTION_ROOT}/project/source0.c" PROPERTIES
  GENERATED TRUE
)
set_source_files_properties("${SOLUTION_ROOT}/project/source1.c" PROPERTIES
  GENERATED TRUE
)
