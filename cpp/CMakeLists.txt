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
        GIT_TAG 3c1f0ad1c44b53d6c17fd7c1d578ec61616b5661
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
    # Guard against circular header in clp/ffi/ir_stream/decoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src
)

target_include_directories(${LIB_NAME}
    PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${CLP_SRC_DIR}/components/core/src
    # CLP source expects string_utils to be included as a library <string_utils/...>, but we
    # directly bundle string_utils source files so we must add it to the include path.
    ${CLP_SRC_DIR}/components/core/src/clp
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
    ${CLP_SRC_DIR}/components/core/src/clp/BufferReader.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/BufferReader.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/Defs.h
    ${CLP_SRC_DIR}/components/core/src/clp/ErrorCode.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/encoding_methods.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/encoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/encoding_methods.inc
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/byteswap.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/encoding_methods.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/encoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/decoding_methods.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/decoding_methods.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/decoding_methods.inc
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/protocol_constants.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/utils.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/ffi/ir_stream/utils.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ir/parsing.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/ir/parsing.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ir/parsing.inc
    ${CLP_SRC_DIR}/components/core/src/clp/ir/types.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/ReaderInterface.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/ReaderInterface.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/string_utils/string_utils.cpp
    ${CLP_SRC_DIR}/components/core/src/clp/string_utils/string_utils.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/TraceableException.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/time_types.hpp
    ${CLP_SRC_DIR}/components/core/src/clp/type_utils.hpp
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
