cmake_minimum_required(VERSION 3.22.1)

project(clp-ffi-go
    VERSION 0.0.2
    DESCRIPTION "clp-ffi-go: A CLP ffi for golang."
    HOMEPAGE_URL https://github.com/y-scope/clp-ffi-go
    LANGUAGES CXX C
)

option(BUILD_SHARED_LIBS "Build using shared libraries" OFF)

set(CMAKE_EXPORT_COMPILE_COMMANDS
    ON
    CACHE BOOL
    "Enable/Disable output of compile commands during generation."
    FORCE
)

if (NOT CMAKE_BUILD_TYPE AND NOT CMAKE_CONFIGURATION_TYPES)
    set(default_build_type "Release")
    message(STATUS "No build type specified. Setting to '${default_build_type}'.")
    set(CMAKE_BUILD_TYPE "${default_build_type}" CACHE STRING "Choose the type of build." FORCE)
endif()

if(clp-ffi-go_IS_TOP_LEVEL)
    # Include dependency settings if the project isn't being included as a subproject.
    # NOTE: We mark the file optional because if the user happens to have the dependencies
    # installed, this file is not necessary.
    include("${CMAKE_SOURCE_DIR}/../build/deps/cmake-settings/all.cmake" OPTIONAL)
endif()

find_package(msgpack-cxx 7.0.0 REQUIRED)
message(STATUS "Found msgpack-cxx ${msgpack-cxx_VERSION}.")

find_package(nlohmann_json REQUIRED)
message(STATUS "Found nlohmann_json ${nlohmann_json_VERSION}.")

# Add ystdlib
set(YSTDLIB_CPP_BUILD_TESTING OFF)
add_subdirectory("${YSTDLIB_SOURCE_DIR}" "${CMAKE_BINARY_DIR}/ystdlib" EXCLUDE_FROM_ALL)

# Setup library name optionally suffixed with the OS and architecture.
set(LIB_NAME "clp-ffi-go" CACHE STRING "Library name optionally containing OS and architecture.")
set(OS_SUFFIX "" CACHE STRING "OS suffix to add to the library name.")
if (NOT OS_SUFFIX STREQUAL "")
    string(APPEND LIB_NAME "_${OS_SUFFIX}")
endif()
set(ARCH_SUFFIX "" CACHE STRING "Architecture suffix to add to the library name.")
if (NOT ARCH_SUFFIX STREQUAL "")
    string(APPEND LIB_NAME "_${ARCH_SUFFIX}")
endif()
add_library(${LIB_NAME})

set_target_properties(${LIB_NAME}
    PROPERTIES
    POSITION_INDEPENDENT_CODE ON
)

# Macro providing the length of the absolute source directory path so we can
# create a relative (rather than absolute) __FILE__ macro
string(LENGTH "${CMAKE_CURRENT_SOURCE_DIR}/" SOURCE_PATH_SIZE)
target_compile_definitions(${LIB_NAME}
    PUBLIC
    SOURCE_PATH_SIZE=${SOURCE_PATH_SIZE}
)

target_compile_features(${LIB_NAME}
    PRIVATE
    cxx_std_20
)

target_compile_options(${LIB_NAME}
    PRIVATE
    $<$<CXX_COMPILER_ID:MSVC>:/W4 /WX>
    $<$<NOT:$<CXX_COMPILER_ID:MSVC>>:-Wall -Wextra -Wpedantic -Werror>
)

target_include_directories(${LIB_NAME}
    PRIVATE
    src
)

# Mark below headers as system headers so that the compiler (including clang-tidy) doesn't generate
# warnings from them.
target_include_directories(${LIB_NAME}
    SYSTEM PRIVATE
    # Since we build CLP source as part of the project we need to put its source directory on the
    # include path to make their include statements work.
    ${CLP_SOURCE_DIR}
    # CLP source code expects string_utils to be included as a library <string_utils/...>, but we
    # directly bundle string_utils source files so we must add it to the include path.
    ${CLP_SOURCE_DIR}/clp
    # CLP puts its submodule dir on its include path and includes json and outcome libraries from
    # their root dir. To replicate this we put the build directory on the include path and
    # dependencies' root directory must be the name CLP expects/uses.
    ${OUTCOME_INCLUDE_DIR}
)

target_link_libraries(${LIB_NAME}
    PRIVATE
    msgpack-cxx
    nlohmann_json::nlohmann_json
)

target_sources(${LIB_NAME}
    PUBLIC
    FILE_SET HEADERS
        BASE_DIRS src/
        FILES
        src/clp_ffi_go/api_decoration.h
        src/clp_ffi_go/defs.h
        src/clp_ffi_go/ir/deserializer.h
        src/clp_ffi_go/ir/serializer.h
    PRIVATE
    ${CLP_SOURCE_DIR}/clp/ffi/SchemaTree.cpp
    ${CLP_SOURCE_DIR}/clp/ir/parsing.cpp
    ${CLP_SOURCE_DIR}/clp/ReaderInterface.cpp
    ${CLP_SOURCE_DIR}/clp/BufferReader.cpp
    ${CLP_SOURCE_DIR}/clp/BufferReader.hpp
    ${CLP_SOURCE_DIR}/clp/Defs.h
    ${CLP_SOURCE_DIR}/clp/ErrorCode.hpp
    ${CLP_SOURCE_DIR}/clp/ReaderInterface.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/encoding_methods.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/encoding_methods.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/encoding_methods.inc
    ${CLP_SOURCE_DIR}/clp/ffi/KeyValuePairLogEvent.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/KeyValuePairLogEvent.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/byteswap.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/Deserializer.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/encoding_methods.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/encoding_methods.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/decoding_methods.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/decoding_methods.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/decoding_methods.inc
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/ir_unit_deserialization_methods.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/ir_unit_deserialization_methods.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/IrUnitHandlerInterface.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/IrUnitType.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/protocol_constants.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/Serializer.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/Serializer.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/utils.cpp
    ${CLP_SOURCE_DIR}/clp/ffi/ir_stream/utils.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/SchemaTree.hpp
    ${CLP_SOURCE_DIR}/clp/ffi/Value.hpp
    ${CLP_SOURCE_DIR}/clp/ir/EncodedTextAst.cpp
    ${CLP_SOURCE_DIR}/clp/ir/EncodedTextAst.hpp
    ${CLP_SOURCE_DIR}/clp/ir/parsing.inc
    ${CLP_SOURCE_DIR}/clp/ir/parsing.hpp
    ${CLP_SOURCE_DIR}/clp/ir/types.hpp
    ${CLP_SOURCE_DIR}/clp/string_utils/string_utils.cpp
    ${CLP_SOURCE_DIR}/clp/string_utils/string_utils.hpp
    ${CLP_SOURCE_DIR}/clp/TraceableException.hpp
    ${CLP_SOURCE_DIR}/clp/time_types.hpp
    ${CLP_SOURCE_DIR}/clp/type_utils.hpp
    src/clp_ffi_go/ir/deserializer.cpp
    src/clp_ffi_go/ir/ErrorCode.hpp
    src/clp_ffi_go/ir/ErrorCode.cpp
    src/clp_ffi_go/ir/serializer.cpp
)

install(TARGETS ${LIB_NAME}
    FILE_SET HEADERS
)
