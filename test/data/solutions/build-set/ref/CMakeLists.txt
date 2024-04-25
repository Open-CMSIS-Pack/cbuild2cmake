cmake_minimum_required(VERSION 3.22)
include(ExternalProject)
	
project("solution" NONE)

# Roots
set(CMSIS_PACK_ROOT "C:/Users/danbro01/AppData/Local/arm/packs")
cmake_path(ABSOLUTE_PATH CMSIS_PACK_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_PACK_ROOT)
set(CMSIS_COMPILER_ROOT "C:/cmsis-toolbox/latest/etc")
cmake_path(ABSOLUTE_PATH CMSIS_COMPILER_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_COMPILER_ROOT)
set(SOLUTION_ROOT "D:/work/cbuild2cmake/test/run/solutions/build-set")
cmake_path(ABSOLUTE_PATH SOLUTION_ROOT NORMALIZE OUTPUT_VARIABLE SOLUTION_ROOT)

# Context specific lists
set(CONTEXTS
  "project.Release+ARMCM0"
)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(DIRS
  "${CMAKE_CURRENT_SOURCE_DIR}/project.Release+ARMCM0"
)

set(OUTPUTS
  "${SOLUTION_ROOT}/out/project/ARMCM0/Release/project.axf"
)

set(ARGS
  "-DSOLUTION_ROOT=${SOLUTION_ROOT}"
  "-DCMSIS_PACK_ROOT=${CMSIS_PACK_ROOT}"
  "-DCMSIS_COMPILER_ROOT=${CMSIS_COMPILER_ROOT}"
)

# Iterate over contexts
foreach(INDEX RANGE ${CONTEXTS_LENGTH})

  math(EXPR N "${INDEX}+1")
  list(GET CONTEXTS ${INDEX} CONTEXT)
  list(GET DIRS ${INDEX} DIR)
  list(GET OUTPUTS ${INDEX} OUTPUT)

  # Create external project, set configure and build steps
  ExternalProject_Add(${CONTEXT}
    PREFIX            ${DIR}
    SOURCE_DIR        ${DIR}
    BINARY_DIR        ${N}
    INSTALL_COMMAND   ""
    TEST_COMMAND      ""
    CONFIGURE_COMMAND ${CMAKE_COMMAND} -G Ninja -S <SOURCE_DIR> -B <BINARY_DIR> ${ARGS} 
    BUILD_COMMAND     ${CMAKE_COMMAND} --build <BINARY_DIR>
    BUILD_ALWAYS      TRUE
    BUILD_BYPRODUCTS  ${OUTPUT}
  )
  ExternalProject_Add_StepTargets(${CONTEXT} build configure)

  # Debug
  message(VERBOSE "Configure Context: ${CMAKE_COMMAND} -G Ninja -S ${DIR} -B ${N}")

  # Database generation step
  ExternalProject_Add_Step(${CONTEXT} database
    COMMAND           ${CMAKE_COMMAND} --build <BINARY_DIR> --target database
    EXCLUDE_FROM_MAIN TRUE
    ALWAYS            TRUE
    DEPENDEES         configure
  )
  ExternalProject_Add_StepTargets(${CONTEXT} database)

endforeach()
