# groups.cmake

# group Group1
add_library(Group_Group1 OBJECT
  "${SOLUTION_ROOT}/project/main.c"
)
target_include_directories(Group_Group1 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Group1 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
add_library(Group_Group1_ABSTRACTIONS INTERFACE)
target_link_libraries(Group_Group1_ABSTRACTIONS INTERFACE
  ${CONTEXT}_ABSTRACTIONS
)
target_compile_options(Group_Group1 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
target_link_libraries(Group_Group1 PUBLIC
  Group_Group1_ABSTRACTIONS
)

# group Group2
add_library(Group_Group2 OBJECT
  "${SOLUTION_ROOT}/project/optimize_none1.c"
  "${SOLUTION_ROOT}/project/optimize_speed1.c"
)
target_include_directories(Group_Group2 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Group2 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
add_library(Group_Group2_ABSTRACTIONS INTERFACE)
cbuild_set_options_flags(CC "none" "off" "" "" CC_OPTIONS_FLAGS_Group_Group2)
target_compile_options(Group_Group2_ABSTRACTIONS INTERFACE
  $<$<COMPILE_LANGUAGE:C>:
    "SHELL:${CC_OPTIONS_FLAGS_Group_Group2}"
  >
)
target_compile_options(Group_Group2 PUBLIC
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)
set(CC_OPTIONS_FLAGS)
cbuild_set_options_flags(CC "none" "off" "" "" CC_OPTIONS_FLAGS)
separate_arguments(CC_OPTIONS_FLAGS)
set_source_files_properties("${SOLUTION_ROOT}/project/optimize_none1.c" PROPERTIES
  COMPILE_OPTIONS "${CC_OPTIONS_FLAGS}"
)
set(CC_OPTIONS_FLAGS)
cbuild_set_options_flags(CC "speed" "off" "" "" CC_OPTIONS_FLAGS)
separate_arguments(CC_OPTIONS_FLAGS)
set_source_files_properties("${SOLUTION_ROOT}/project/optimize_speed1.c" PROPERTIES
  COMPILE_OPTIONS "${CC_OPTIONS_FLAGS}"
)

# group SubGroup
add_library(Group_Group2_SubGroup OBJECT
  "${SOLUTION_ROOT}/project/optimize_none2.c"
)
target_include_directories(Group_Group2_SubGroup PUBLIC
  $<TARGET_PROPERTY:Group_Group2,INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Group2_SubGroup PUBLIC
  $<TARGET_PROPERTY:Group_Group2,INTERFACE_COMPILE_DEFINITIONS>
)
add_library(Group_Group2_SubGroup_ABSTRACTIONS INTERFACE)
target_link_libraries(Group_Group2_SubGroup_ABSTRACTIONS INTERFACE
  Group_Group2_ABSTRACTIONS
)
target_compile_options(Group_Group2_SubGroup PUBLIC
  $<TARGET_PROPERTY:Group_Group2,INTERFACE_COMPILE_OPTIONS>
)
target_link_libraries(Group_Group2_SubGroup PUBLIC
  Group_Group2_SubGroup_ABSTRACTIONS
)

# group SubGroup2
add_library(Group_Group2_SubGroup_SubGroup2 OBJECT
  "${SOLUTION_ROOT}/project/optimize_speed2.c"
)
target_include_directories(Group_Group2_SubGroup_SubGroup2 PUBLIC
  $<TARGET_PROPERTY:Group_Group2_SubGroup,INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_Group2_SubGroup_SubGroup2 PUBLIC
  $<TARGET_PROPERTY:Group_Group2_SubGroup,INTERFACE_COMPILE_DEFINITIONS>
)
add_library(Group_Group2_SubGroup_SubGroup2_ABSTRACTIONS INTERFACE)
cbuild_set_options_flags(CC "speed" "off" "" "" CC_OPTIONS_FLAGS_Group_Group2_SubGroup_SubGroup2)
target_compile_options(Group_Group2_SubGroup_SubGroup2_ABSTRACTIONS INTERFACE
  $<$<COMPILE_LANGUAGE:C>:
    "SHELL:${CC_OPTIONS_FLAGS_Group_Group2_SubGroup_SubGroup2}"
  >
)
target_compile_options(Group_Group2_SubGroup_SubGroup2 PUBLIC
  $<TARGET_PROPERTY:Group_Group2_SubGroup,INTERFACE_COMPILE_OPTIONS>
)
target_link_libraries(Group_Group2_SubGroup_SubGroup2 PUBLIC
  Group_Group2_SubGroup_SubGroup2_ABSTRACTIONS
)

# group EmptyParent
add_library(Group_EmptyParent INTERFACE)
target_include_directories(Group_EmptyParent INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_EmptyParent INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_DEFINITIONS>
)
add_library(Group_EmptyParent_ABSTRACTIONS INTERFACE)
target_link_libraries(Group_EmptyParent_ABSTRACTIONS INTERFACE
  ${CONTEXT}_ABSTRACTIONS
)
target_compile_options(Group_EmptyParent INTERFACE
  $<TARGET_PROPERTY:${CONTEXT},INTERFACE_COMPILE_OPTIONS>
)

# group NestedChild
add_library(Group_EmptyParent_NestedChild OBJECT
  "${SOLUTION_ROOT}/project/optimize_size1.c"
  "${SOLUTION_ROOT}/project/optimize_size2.c"
)
target_include_directories(Group_EmptyParent_NestedChild PUBLIC
  $<TARGET_PROPERTY:Group_EmptyParent,INTERFACE_INCLUDE_DIRECTORIES>
)
target_compile_definitions(Group_EmptyParent_NestedChild PUBLIC
  $<TARGET_PROPERTY:Group_EmptyParent,INTERFACE_COMPILE_DEFINITIONS>
)
add_library(Group_EmptyParent_NestedChild_ABSTRACTIONS INTERFACE)
target_link_libraries(Group_EmptyParent_NestedChild_ABSTRACTIONS INTERFACE
  Group_EmptyParent_ABSTRACTIONS
)
target_compile_options(Group_EmptyParent_NestedChild PUBLIC
  $<TARGET_PROPERTY:Group_EmptyParent,INTERFACE_COMPILE_OPTIONS>
)
target_link_libraries(Group_EmptyParent_NestedChild PUBLIC
  Group_EmptyParent_NestedChild_ABSTRACTIONS
)
