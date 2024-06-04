cmake_minimum_required(VERSION 3.23)
include(FetchContent)

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

# Setup library name based on Go environment variables set by `go generate`
set(LIB_NAME "clp_ffi" CACHE STRING "Library name containing os and arch.")
if (DEFINED ENV{GOOS})
    string(APPEND LIB_NAME "_$ENV{GOOS}")
endif()
if (DEFINED ENV{GOARCH})
    string(APPEND LIB_NAME "_$ENV{GOARCH}")
endif()

set(CLP_SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}/clp" CACHE STRING
    "Directory containing CLP source, fetched from github if non existant.")
if (NOT EXISTS ${CLP_SRC_DIR})
    FetchContent_Declare(
        clp-core
        GIT_REPOSITORY https://github.com/y-scope/clp.git
        GIT_TAG 084efa35b7e9a63aecc5e327b97aea2a1cef83bc
        SOURCE_DIR ${CLP_SRC_DIR}
    )
    message(STATUS "Fetching CLP from github.")
    FetchContent_MakeAvailable(clp-core)
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
    SYSTEM PRIVATE
    ${CLP_SRC_DIR}/components/core/submodules
    ${CMAKE_CURRENT_SOURCE_DIR}
)

target_include_directories(${LIB_NAME}
    PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
)

target_sources(${LIB_NAME} PUBLIC
    PUBLIC
    FILE_SET HEADERS
        BASE_DIRS src/
        FILES
        src/ffi_go/api_decoration.h
        src/ffi_go/defs.h
        src/ffi_go/ir/decoder.h
        src/ffi_go/ir/deserializer.h
        src/ffi_go/ir/encoder.h
        src/ffi_go/ir/serializer.h
        src/ffi_go/search/wildcard_query.h
    PRIVATE
    ${CLP_SRC_DIR}/components/core/src/BufferReader.cpp
    ${CLP_SRC_DIR}/components/core/src/BufferReader.hpp
    ${CLP_SRC_DIR}/components/core/src/Defs.h
    ${CLP_SRC_DIR}/components/core/src/ErrorCode.hpp
    ${CLP_SRC_DIR}/components/core/src/ReaderInterface.cpp
    ${CLP_SRC_DIR}/components/core/src/ReaderInterface.hpp
    ${CLP_SRC_DIR}/components/core/src/string_utils.cpp
    ${CLP_SRC_DIR}/components/core/src/string_utils.hpp
    ${CLP_SRC_DIR}/components/core/src/string_utils.inc
    ${CLP_SRC_DIR}/components/core/src/TraceableException.hpp
    ${CLP_SRC_DIR}/components/core/src/type_utils.hpp
    ${CLP_SRC_DIR}/components/core/src/ffi/encoding_methods.cpp
    ${CLP_SRC_DIR}/components/core/src/ffi/encoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src/ffi/encoding_methods.inc
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/byteswap.hpp
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/encoding_methods.cpp
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/encoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/decoding_methods.cpp
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/decoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/decoding_methods.inc
    ${CLP_SRC_DIR}/components/core/src/ffi/ir_stream/protocol_constants.hpp
    src/ffi_go/types.hpp
    src/ffi_go/ir/decoder.cpp
    src/ffi_go/ir/deserializer.cpp
    src/ffi_go/ir/encoder.cpp
    src/ffi_go/ir/types.hpp
    src/ffi_go/ir/serializer.cpp
    src/ffi_go/search/wildcard_query.cpp
)

include(GNUInstallDirs)
install(TARGETS ${LIB_NAME}
    ARCHIVE
    FILE_SET HEADERS
)
