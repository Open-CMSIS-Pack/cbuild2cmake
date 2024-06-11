# groups.cmake

# group Source1
add_library(Group_Source1 OBJECT
  "${SOLUTION_ROOT}/project/source1.c"
)
target_include_directories(Group_Source1 PUBLIC
  $<$<COMPILE_LANGUAGE:ASM>:
    ${SOLUTION_ROOT}/project/group
  >
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ${SOLUTION_ROOT}/project/inc1
  >
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Source1 PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    DEF1=1
  >
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(Group_Source1 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)

# file asm.s
add_library(Group_Source1_asm_s OBJECT
  "${SOLUTION_ROOT}/project/asm.s"
)
target_include_directories(Group_Source1_asm_s PUBLIC
  $<$<COMPILE_LANGUAGE:ASM>:
    ${SOLUTION_ROOT}/project/file
  >
  $<TARGET_PROPERTY:Group_Source1,INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_options(Group_Source1_asm_s PUBLIC
  $<TARGET_PROPERTY:Group_Source1,INTERFACE_COMPILE_OPTIONS>
)
set(COMPILE_DEFINITIONS
  ARMCM0
  _RTE_
)
cbuild_set_defines(AS_GNU COMPILE_DEFINITIONS)
set_source_files_properties("${SOLUTION_ROOT}/project/asm.s" PROPERTIES
  COMPILE_FLAGS "${COMPILE_DEFINITIONS}"
)

# file source3.c
add_library(Group_Source1_source3_c OBJECT
  "${SOLUTION_ROOT}/project/source3.c"
)
target_include_directories(Group_Source1_source3_c PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ${SOLUTION_ROOT}/project/inc3
  >
  $<LIST:REMOVE_ITEM,$<TARGET_PROPERTY:Group_Source1,INTERFACE_INCLUDE_DIRECTORIES>,${SOLUTION_ROOT}/project/inc1>
)
target_compile_definitions(Group_Source1_source3_c PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    DEF3
  >
  $<LIST:FILTER,$<TARGET_PROPERTY:Group_Source1,INTERFACE_COMPILE_DEFINITIONS>,EXCLUDE,^DEF1.*>
)
target_compile_options(Group_Source1_source3_c PUBLIC
  $<TARGET_PROPERTY:Group_Source1,INTERFACE_COMPILE_OPTIONS>
)

# group Source2
add_library(Group_Source1_Source2 OBJECT
  "${SOLUTION_ROOT}/project/source2.c"
)
target_include_directories(Group_Source1_Source2 PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ${SOLUTION_ROOT}/project/inc2
  >
  $<LIST:REMOVE_ITEM,$<TARGET_PROPERTY:Group_Source1,INTERFACE_INCLUDE_DIRECTORIES>,${SOLUTION_ROOT}/project/inc1>
)
target_compile_definitions(Group_Source1_Source2 PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    DEF2=1
  >
  $<LIST:FILTER,$<TARGET_PROPERTY:Group_Source1,INTERFACE_COMPILE_DEFINITIONS>,EXCLUDE,^DEF1.*>
)
target_compile_options(Group_Source1_Source2 PUBLIC
  $<TARGET_PROPERTY:Group_Source1,INTERFACE_COMPILE_OPTIONS>
)

# group Main
add_library(Group_Main INTERFACE)
target_include_directories(Group_Main INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Main INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(Group_Main INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)

# file main.c
add_library(Group_Main_main_c OBJECT
  "${SOLUTION_ROOT}/project/main.c"
)
target_include_directories(Group_Main_main_c PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    ${SOLUTION_ROOT}/project/inc2
  >
  $<TARGET_PROPERTY:Group_Main,INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Main_main_c PUBLIC
  $<$<COMPILE_LANGUAGE:C,CXX>:
    DEF2
  >
  $<TARGET_PROPERTY:Group_Main,INTERFACE_COMPILE_DEFINITIONS>
)
target_compile_options(Group_Main_main_c PUBLIC
  $<TARGET_PROPERTY:Group_Main,INTERFACE_COMPILE_OPTIONS>
)

# group Headers
add_library(Group_Headers INTERFACE)
target_include_directories(Group_Headers INTERFACE
  ${SOLUTION_ROOT}/project/inc3
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Headers INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
