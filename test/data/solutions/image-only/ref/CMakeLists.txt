cmake_minimum_required(VERSION 3.27)

project("solution" NONE)

# Roots
include("roots.cmake")

# Execute: Convert_Image1
set(INPUT
  ${SOLUTION_ROOT}/images/image1.elf
)
list(GET INPUT 0 INPUT_0)
set(OUTPUT
  ${SOLUTION_ROOT}/images/image2.hex
)
list(GET OUTPUT 0 OUTPUT_0)
add_custom_target(Convert_Image1 ALL DEPENDS ${OUTPUT})
add_custom_command(OUTPUT ${OUTPUT} DEPENDS ${INPUT}
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --green "Executing: Convert_Image1"
  COMMAND ${CMAKE_COMMAND} -E echo "Simulate image conversion 1 > 2" && ${CMAKE_COMMAND} -E copy "${INPUT_0}" "${OUTPUT_0}"
  USES_TERMINAL
)

# Execute: Convert_Image2
set(INPUT
  ${SOLUTION_ROOT}/images/image2.hex
)
list(GET INPUT 0 INPUT_0)
set(OUTPUT
  ${SOLUTION_ROOT}/images/image3.bin
)
list(GET OUTPUT 0 OUTPUT_0)
add_custom_target(Convert_Image2 ALL DEPENDS ${OUTPUT})
add_custom_command(OUTPUT ${OUTPUT} DEPENDS ${INPUT}
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --green "Executing: Convert_Image2"
  COMMAND ${CMAKE_COMMAND} -E echo "Simulate image conversion 2 > 3" && ${CMAKE_COMMAND} -E copy "${INPUT_0}" "${OUTPUT_0}"
  USES_TERMINAL
)

# Build dependencies
add_dependencies(Convert_Image2
  Convert_Image1
)
