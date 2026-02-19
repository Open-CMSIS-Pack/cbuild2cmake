cmake_minimum_required(VERSION 3.27)
include(ExternalProject)
	
project("solution" NONE)

# Enable color diagnostics
set(CMAKE_COLOR_DIAGNOSTICS ON)

# Roots
include("roots.cmake")

# Context specific lists
set(CONTEXTS
  "core0.Debug+CM0"
  "core1.Debug+CM0"
)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(COMPILERS
  "AC6 V6.19.0"
  "AC6 V6.19.0"
)

set(DIRS
  "${CMAKE_CURRENT_SOURCE_DIR}/core0.Debug+CM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/core1.Debug+CM0"
)

set(WEST_CONTEXTS
  "true"
  "true"
)

set(OUTPUTS_1
  "${SOLUTION_ROOT}/out/core0/CM0/Debug/zephyr/zephyr.elf"
  "${SOLUTION_ROOT}/out/core0/CM0/Debug/zephyr/zephyr.hex"
)
set(OUTPUTS_2
  "${SOLUTION_ROOT}/out/core1/CM0/Debug/zephyr/zephyr.elf"
  "${SOLUTION_ROOT}/out/core1/CM0/Debug/zephyr/zephyr.hex"
)

set(ARGS
  "-DSOLUTION_ROOT=${SOLUTION_ROOT}"
  "-DCMSIS_PACK_ROOT=${CMSIS_PACK_ROOT}"
  "-DCMSIS_COMPILER_ROOT=${CMSIS_COMPILER_ROOT}"
)

# Compilation database
add_custom_target(database)

# Iterate over contexts
foreach(INDEX RANGE ${CONTEXTS_LENGTH})

  math(EXPR N "${INDEX}+1")
  list(GET CONTEXTS ${INDEX} CONTEXT)
  list(GET COMPILERS ${INDEX} COMPILER)
  list(GET DIRS ${INDEX} DIR)
  list(GET WEST_CONTEXTS ${INDEX} WEST_CONTEXT)
  if(WEST_CONTEXT)
    set(WEST_TARGET "--target west")
  endif()

  # Create external project, set configure and build steps
  ExternalProject_Add(${CONTEXT}
    PREFIX                ${DIR}
    SOURCE_DIR            ${DIR}
    BINARY_DIR            ${N}
    INSTALL_COMMAND       ""
    TEST_COMMAND          ""
    CONFIGURE_COMMAND     ${CMAKE_COMMAND} -G Ninja -S <SOURCE_DIR> -B <BINARY_DIR> ${ARGS} 
    BUILD_COMMAND         ${CMAKE_COMMAND} -E cmake_echo_color --blue --bold "Building CMake target '${CONTEXT}'"
    COMMAND               ${CMAKE_COMMAND} -E echo "Using compiler: ${COMPILER}"
    COMMAND               ${CMAKE_COMMAND} --build <BINARY_DIR> ${WEST_TARGET} --verbose
    BUILD_ALWAYS          TRUE
    BUILD_BYPRODUCTS      ${OUTPUTS_${N}}
    USES_TERMINAL_BUILD   ON
  )

  # Executes command step
  ExternalProject_Add_Step(${CONTEXT} executes
    DEPENDEES         build
  )

  ExternalProject_Add_StepTargets(${CONTEXT} build configure executes)

  # Debug
  message(VERBOSE "Configure Context: ${CMAKE_COMMAND} -G Ninja -S ${DIR} -B ${N}")

  # Database generation step
  ExternalProject_Add_Step(${CONTEXT} database
    COMMAND           ${CMAKE_COMMAND} --build <BINARY_DIR> --target database --verbose
    ALWAYS            TRUE
    USES_TERMINAL     ON
    DEPENDEES         configure
  )
  ExternalProject_Add_StepTargets(${CONTEXT} database)
  add_dependencies(database ${CONTEXT}-database)

endforeach()
