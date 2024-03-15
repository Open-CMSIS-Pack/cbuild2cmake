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

# group IAR
add_library(Group_IAR OBJECT
  "${SOLUTION_ROOT}/project/IAR/Asm.s"
)
target_link_libraries(Group_IAR PRIVATE
  ${CONTEXT}_GLOBAL
  ${CONTEXT}_INCLUDES
  ${CONTEXT}_DEFINES
)
set_source_files_properties("${SOLUTION_ROOT}/project/IAR/Asm.s" PROPERTIES
  COMPILE_DEFINITIONS IAR_ASM_DEF
)
