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

# group AC6
add_library(Group_AC6 OBJECT
  "${SOLUTION_ROOT}/project/AC6/AsmArm.s"
  "${SOLUTION_ROOT}/project/AC6/GnuSyntax.s"
  "${SOLUTION_ROOT}/project/AC6/PreProcessed.S"
  "${SOLUTION_ROOT}/project/AC6/Auto.s"
)
target_link_libraries(Group_AC6 PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)
set(COMPILE_DEFINITIONS
  HEXADECIMAL_TEST=11259375
  DECIMAL_TEST=1234567890
  STRING_TEST="String0"
)
cbuild_set_defines(AS_ARM COMPILE_DEFINITIONS)
set_source_files_properties("${SOLUTION_ROOT}/project/AC6/AsmArm.s" PROPERTIES
  COMPILE_FLAGS "${COMPILE_DEFINITIONS}"
)
set(COMPILE_DEFINITIONS
  GAS_DEF
)
cbuild_set_defines(AS_GNU COMPILE_DEFINITIONS)
set_source_files_properties("${SOLUTION_ROOT}/project/AC6/GnuSyntax.s" PROPERTIES
  COMPILE_FLAGS "${COMPILE_DEFINITIONS}"
)
set_source_files_properties("${SOLUTION_ROOT}/project/AC6/GnuSyntax.s" PROPERTIES
  COMPILE_OPTIONS -masm=gnu
)
set_source_files_properties("${SOLUTION_ROOT}/project/AC6/PreProcessed.S" PROPERTIES
  COMPILE_DEFINITIONS PRE_PROCESSED_DEF
)
set(COMPILE_DEFINITIONS
  AUTO_DEF
)
cbuild_set_defines(AS_ARM COMPILE_DEFINITIONS)
set_source_files_properties("${SOLUTION_ROOT}/project/AC6/Auto.s" PROPERTIES
  COMPILE_FLAGS "${COMPILE_DEFINITIONS}"
)
