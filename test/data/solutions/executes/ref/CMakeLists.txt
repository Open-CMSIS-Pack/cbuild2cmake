cmake_minimum_required(VERSION 3.22)
include(ExternalProject)
	
project("solution" NONE)

# Roots
set(CMSIS_PACK_ROOT "C:/Users/danbro01/AppData/Local/arm/packs")
cmake_path(ABSOLUTE_PATH CMSIS_PACK_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_PACK_ROOT)
set(CMSIS_COMPILER_ROOT "D:/work/cbuild2cmake/test/run/etc")
cmake_path(ABSOLUTE_PATH CMSIS_COMPILER_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_COMPILER_ROOT)
set(SOLUTION_ROOT "D:/work/cbuild2cmake/test/run/solutions/executes")
cmake_path(ABSOLUTE_PATH SOLUTION_ROOT NORMALIZE OUTPUT_VARIABLE SOLUTION_ROOT)

# Context specific lists
set(CONTEXTS
  "project.Debug+ARMCM0"
  "project.Release+ARMCM0"
)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(DIRS
  "${CMAKE_CURRENT_SOURCE_DIR}/project.Debug+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.Release+ARMCM0"
)

set(OUTPUTS
  "${SOLUTION_ROOT}/out/project/ARMCM0/Debug/project.axf"
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
    BUILD_COMMAND     ${CMAKE_COMMAND} --build <BINARY_DIR> --verbose
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

# Execute: Archive_Artifacts
set(INPUT
  ${SOLUTION_ROOT}/script/archive.cmake
  ${SOLUTION_ROOT}/out/project/ARMCM0/Release/project.axf
  ${SOLUTION_ROOT}/out/project/ARMCM0/Release/project.axf.signed
)
list(GET INPUT 0 INPUT_0)
set(OUTPUT
  ${SOLUTION_ROOT}/artifacts.zip
)
add_custom_target(Archive_Artifacts ALL DEPENDS ${OUTPUT})
add_custom_command(OUTPUT ${OUTPUT} DEPENDS ${INPUT}
  COMMAND ${CMAKE_COMMAND} -DINPUT="${INPUT}" -DOUTPUT="${OUTPUT}" -P "${INPUT_0}"
)

# Execute: Generate_Project_Sources
set(INPUT
  ${SOLUTION_ROOT}/script/generate-sources.cmake
  ${SOLUTION_ROOT}/project/source.c.template
)
list(GET INPUT 0 INPUT_0)
list(GET INPUT 1 INPUT_1)
set(OUTPUT
  ${SOLUTION_ROOT}/project/source0.c
  ${SOLUTION_ROOT}/project/source1.c
)
list(GET OUTPUT 0 OUTPUT_0)
list(GET OUTPUT 1 OUTPUT_1)
add_custom_target(Generate_Project_Sources ALL DEPENDS ${OUTPUT})
add_custom_command(OUTPUT ${OUTPUT} DEPENDS ${INPUT}
  COMMAND ${CMAKE_COMMAND} -DINPUT_1="${INPUT_1}" -DOUTPUT_0="${OUTPUT_0}" -DOUTPUT_1="${OUTPUT_1}" -P "${INPUT_0}"
)

# Execute: Run_After_Archiving
set(INPUT
  ${SOLUTION_ROOT}/artifacts.zip
)
set(OUTPUT
  ${CMAKE_CURRENT_BINARY_DIR}/Run_After_Archiving.stamp
)
add_custom_target(Run_After_Archiving ALL DEPENDS ${OUTPUT})
add_custom_command(OUTPUT ${OUTPUT} DEPENDS ${INPUT}
  COMMAND ${CMAKE_COMMAND} -E touch "Run_After_Archiving.stamp"
  COMMAND ${CMAKE_COMMAND} -E echo "Archive has been updated"
)

# Execute: Run_Always
add_custom_target(Run_Always ALL
  COMMAND ${CMAKE_COMMAND} -E echo "Execute Run Always"
)

# Execute: project.Release+ARMCM0-Sign_Artifact
set(INPUT
  ${SOLUTION_ROOT}/script/sign.cmake
  ${SOLUTION_ROOT}/out/project/ARMCM0/Release/project.axf
)
list(GET INPUT 0 INPUT_0)
set(OUTPUT
  ${SOLUTION_ROOT}/out/project/ARMCM0/Release/project.axf.signed
)
add_custom_target(project.Release+ARMCM0-Sign_Artifact ALL DEPENDS ${OUTPUT})
add_custom_command(OUTPUT ${OUTPUT} DEPENDS ${INPUT}
  COMMAND ${CMAKE_COMMAND} -DINPUT="${INPUT}" -DOUTPUT="${OUTPUT}" -P "${INPUT_0}"
)

# Build dependencies
add_dependencies(project.Debug+ARMCM0-build
  Generate_Project_Sources
)
add_dependencies(project.Release+ARMCM0-build
  Generate_Project_Sources
)
add_dependencies(Archive_Artifacts
  project.Release+ARMCM0-build
  project.Release+ARMCM0-Sign_Artifact
)
add_dependencies(Run_After_Archiving
  Archive_Artifacts
)
add_dependencies(project.Release+ARMCM0-Sign_Artifact
  project.Release+ARMCM0-build
)
