cmake_minimum_required(VERSION 3.27)
include(ExternalProject)
	
project("solution" NONE)

# Roots
include("roots.cmake")

# Context specific lists
set(CONTEXTS
  "project.AC6+ARMCM0"
  "project.CLANG+ARMCM0"
  "project.GCC+ARMCM0"
  "project.IAR+ARMCM0"
)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(DIRS
  "${CMAKE_CURRENT_SOURCE_DIR}/project.AC6+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.CLANG+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.GCC+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.IAR+ARMCM0"
)

set(OUTPUTS_1
  "${SOLUTION_ROOT}/../pre-include-oot/out/project/ARMCM0/AC6/project.axf"
)
set(OUTPUTS_2
  "${SOLUTION_ROOT}/../pre-include-oot/out/project/ARMCM0/CLANG/project.elf"
)
set(OUTPUTS_3
  "${SOLUTION_ROOT}/../pre-include-oot/out/project/ARMCM0/GCC/project.elf"
)
set(OUTPUTS_4
  "${SOLUTION_ROOT}/../pre-include-oot/out/project/ARMCM0/IAR/project.out"
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
  list(GET DIRS ${INDEX} DIR)

  # Create external project, set configure and build steps
  ExternalProject_Add(${CONTEXT}
    PREFIX                ${DIR}
    SOURCE_DIR            ${DIR}
    BINARY_DIR            ${N}
    INSTALL_COMMAND       ""
    TEST_COMMAND          ""
    CONFIGURE_COMMAND     ${CMAKE_COMMAND} -G Ninja -S <SOURCE_DIR> -B <BINARY_DIR> ${ARGS} 
    BUILD_COMMAND         ${CMAKE_COMMAND} -E echo "Building CMake target '${CONTEXT}'"
    COMMAND               ${CMAKE_COMMAND} --build <BINARY_DIR> --verbose
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
    COMMAND           ${CMAKE_COMMAND} --build <BINARY_DIR> --target database
    ALWAYS            TRUE
    DEPENDEES         configure
  )
  ExternalProject_Add_StepTargets(${CONTEXT} database)
  add_dependencies(database ${CONTEXT}-database)

endforeach()
