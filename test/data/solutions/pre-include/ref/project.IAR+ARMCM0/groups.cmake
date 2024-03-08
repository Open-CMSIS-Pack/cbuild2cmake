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
