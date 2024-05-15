# groups.cmake

# group Source
add_library(Group_Source OBJECT
  "${SOLUTION_ROOT}/project/main.c"
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

# group GCC-CLANG
add_library(Group_GCC-CLANG OBJECT
  "${SOLUTION_ROOT}/project/GCC/PreProcessed.S"
  "${SOLUTION_ROOT}/project/GCC/NonPreProcessed.s"
)
target_include_directories(Group_GCC-CLANG PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_GCC-CLANG PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(Group_GCC-CLANG PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
set_source_files_properties("${SOLUTION_ROOT}/project/GCC/PreProcessed.S" PROPERTIES
  COMPILE_DEFINITIONS "PRE_PROCESSED_DEF;GROUP_ASM_GCC_CLANG_DEF"
)
set_source_files_properties("${SOLUTION_ROOT}/project/GCC/NonPreProcessed.s" PROPERTIES
  COMPILE_DEFINITIONS "GROUP_ASM_GCC_CLANG_DEF"
)
set_source_files_properties("${SOLUTION_ROOT}/project/GCC/NonPreProcessed.s" PROPERTIES
  COMPILE_OPTIONS "-DPRE_PROCESSED_DEF"
)
