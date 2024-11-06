cmake_minimum_required(VERSION 3.22.1)

# Turn on verbose output
set(CMAKE_VERBOSE_MAKEFILE ON)

project(clp_ffi
    LANGUAGES CXX C
)

# Enable compile commands by default if the generator supports it.
if (NOT CMAKE_EXPORT_COMPILE_COMMANDS AND CMAKE_GENERATOR MATCHES "Ninja|Unix Makefiles")
    set(CMAKE_EXPORT_COMPILE_COMMANDS ON CACHE BOOL
        "Enable/Disable output of compile commands during generation." FORCE)
endif()

# Set default build type to Release
if (NOT CMAKE_BUILD_TYPE AND NOT CMAKE_CONFIGURATION_TYPES)
    set(default_build_type "Release")
    message(STATUS "No build type specified. Setting to '${default_build_type}'.")
    set(CMAKE_BUILD_TYPE "${default_build_type}" CACHE STRING "Choose the type of build." FORCE)
endif()

# Build/package static by default to simplify compatibility in other systems
option(BUILD_SHARED_LIBS "Build using shared libraries" OFF)

# Setup library name optionally suffixed with the OS and architecture.
set(LIB_NAME "clp_ffi" CACHE STRING "Library name optionally containing OS and architecture.")
set(OS_SUFFIX "" CACHE STRING "OS suffix to add to the library name.")
if (NOT OS_SUFFIX STREQUAL "")
    string(APPEND LIB_NAME "_${OS_SUFFIX}")
endif()
set(ARCH_SUFFIX "" CACHE STRING "Architecture suffix to add to the library name.")
if (NOT ARCH_SUFFIX STREQUAL "")
    string(APPEND LIB_NAME "_${ARCH_SUFFIX}")
endif()

# TODO: remove deps_dir/clp/json once we can build and install clp-ffi-core as a library.
set(DEPS_DIR "../build/deps" CACHE STRING "Path dependency directory.")
set(clp_SOURCE_DIR
    "${DEPS_DIR}/clp-src/clp/components/core/src"
    CACHE STRING
    "Path to CLP core source directory.")
set(json_SOURCE_DIR "${DEPS_DIR}/json-src" CACHE STRING "Path to outcome source directory.")
set(outcome_SOURCE_DIR "${DEPS_DIR}/outcome-src" CACHE STRING "Path to outcome source directory.")
find_package(msgpack-cxx 6.1.0 REQUIRED)
if(msgpack-cxx_FOUND)
    message(STATUS "Found msgpack-cxx ${msgpack-cxx_VERSION}")
else()
    message(FATAL_ERROR "Could not find msgpack-cxx")
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

# Set warnings as errors
target_compile_options(${LIB_NAME}
    PRIVATE
    $<$<CXX_COMPILER_ID:MSVC>:/W4 /WX>
    $<$<NOT:$<CXX_COMPILER_ID:MSVC>>:-Wall -Wextra -Wpedantic -Werror>
)

target_include_directories(${LIB_NAME}
    PRIVATE
    src
)

target_link_libraries(${LIB_NAME}
    PRIVATE
    msgpack-cxx
)

# Mark below headers as system headers so that the compiler (including clang-tidy) doesn't generate
# warnings from them.
target_include_directories(${LIB_NAME}
    SYSTEM PRIVATE
    ${outcome_SOURCE_DIR}

    # Since we build CLP source as part of the project we need to put its source directory on the
    # include path to make their include statements work.
    ${clp_SOURCE_DIR}
    # CLP source code expects string_utils to be included as a library <string_utils/...>, but we
    # directly bundle string_utils source files so we must add it to the include path.
    ${clp_SOURCE_DIR}/clp
    # CLP puts its submodule dir on its include path and includes json and outcome libraries from
    # their root dir. To replicate this we put the build directory on the include path and
    # dependencies' root directory must be the name CLP expects/uses.
    ${json_SOURCE_DIR}
)

target_sources(${LIB_NAME}
    PUBLIC
    FILE_SET HEADERS
        BASE_DIRS src/
        FILES
        src/ffi_go/api_decoration.h
        src/ffi_go/defs.h
        src/ffi_go/ir/deserializer.h
        src/ffi_go/ir/serializer.h
    PRIVATE
    ${clp_SOURCE_DIR}/clp/ffi/SchemaTree.cpp
    ${clp_SOURCE_DIR}/clp/ir/parsing.cpp
    ${clp_SOURCE_DIR}/clp/ReaderInterface.cpp
    ${clp_SOURCE_DIR}/clp/BufferReader.cpp
    ${clp_SOURCE_DIR}/clp/BufferReader.hpp
    ${clp_SOURCE_DIR}/clp/Defs.h
    ${clp_SOURCE_DIR}/clp/ErrorCode.hpp
    ${clp_SOURCE_DIR}/clp/ReaderInterface.hpp
    ${clp_SOURCE_DIR}/clp/ffi/encoding_methods.cpp
    ${clp_SOURCE_DIR}/clp/ffi/encoding_methods.hpp
    ${clp_SOURCE_DIR}/clp/ffi/encoding_methods.inc
    ${clp_SOURCE_DIR}/clp/ffi/KeyValuePairLogEvent.cpp
    ${clp_SOURCE_DIR}/clp/ffi/KeyValuePairLogEvent.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/byteswap.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/Deserializer.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/encoding_methods.cpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/encoding_methods.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/decoding_methods.cpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/decoding_methods.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/decoding_methods.inc
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/ir_unit_deserialization_methods.cpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/ir_unit_deserialization_methods.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/IrUnitHandlerInterface.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/IrUnitType.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/protocol_constants.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/Serializer.cpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/Serializer.hpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/utils.cpp
    ${clp_SOURCE_DIR}/clp/ffi/ir_stream/utils.hpp
    ${clp_SOURCE_DIR}/clp/ffi/SchemaTree.hpp
    ${clp_SOURCE_DIR}/clp/ffi/Value.hpp
    ${clp_SOURCE_DIR}/clp/ir/EncodedTextAst.cpp
    ${clp_SOURCE_DIR}/clp/ir/EncodedTextAst.hpp
    ${clp_SOURCE_DIR}/clp/ir/parsing.inc
    ${clp_SOURCE_DIR}/clp/ir/parsing.hpp
    ${clp_SOURCE_DIR}/clp/ir/types.hpp
    ${clp_SOURCE_DIR}/clp/string_utils/string_utils.cpp
    ${clp_SOURCE_DIR}/clp/string_utils/string_utils.hpp
    ${clp_SOURCE_DIR}/clp/TraceableException.hpp
    ${clp_SOURCE_DIR}/clp/time_types.hpp
    ${clp_SOURCE_DIR}/clp/type_utils.hpp
    src/ffi_go/ir/deserializer.cpp
    src/ffi_go/ir/serializer.cpp
    src/lint/msgpack.hpp
)

install(TARGETS ${LIB_NAME}
    FILE_SET HEADERS
)
