cmake_minimum_required(VERSION 3.30)
include(FetchContent)

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
set(OS_SUFFIX"" CACHE STRING "OS suffix to add to the library name.")
string(APPEND LIB_NAME "${OS_SUFFIX}")
set(ARCH_SUFFIX "" CACHE STRING "Architecture suffix to add to the library name.")
string(APPEND LIB_NAME "${ARCH_SUFFIX}")

set(clp_SOURCE_DIR "../build/deps/clp" CACHE STRING "Path to CLP source directory.")
set(outcome_SOURCE_DIR "../build/deps/outcome" CACHE STRING "Path to outcome source directory.")
# set(nlohmann_json_ROOT "../build/deps/json-install")
set(nlohmann_json_DIR "../build/deps/json-install/share/cmake/nlohmann_json")
find_package(nlohmann_json 3.11.3 REQUIRED)
if(nlohmann_json_FOUND)
    message(STATUS "Found nlohmann_json ${nlohmann_json_VERSION}")
else()
    message(FATAL_ERROR "Could not find nlohmann_json")
endif()
set(msgpack-cxx_DIR "../build/deps/msgpack-install/lib/cmake/msgpack-cxx")
find_package(msgpack-cxx 6.2.0 REQUIRED)
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
    nlohmann_json::nlohmann_json
    msgpack-cxx
)

# Mark below headers as system headers so that the compiler (including clang-tidy) doesn't generate
# warnings from them.
target_include_directories(${LIB_NAME}
    SYSTEM PRIVATE
    ${clp_SOURCE_DIR}/components/core/src

    # CLP source code expects string_utils to be included as a library <string_utils/...>, but we
    # directly bundle string_utils source files so we must add it to the include path.
    ${clp_SOURCE_DIR}/components/core/src/clp

    # CLP puts its submodule dir on its include path and includes json and outcome libraries from
    # their root dir. To replicate this we need to put the build path on the include path.
    ${nlohmann_json_INCLUDE_DIRS}
    ${msgpack-cxx_INCLUDE_DIRS}
    ${outcome_SOURCE_DIR}
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
    ${clp_SOURCE_DIR}/components/core/src/clp/BufferReader.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/BufferReader.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/Defs.h
    ${clp_SOURCE_DIR}/components/core/src/clp/ErrorCode.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/encoding_methods.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/encoding_methods.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/encoding_methods.inc
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/byteswap.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/Deserializer.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/encoding_methods.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/encoding_methods.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/decoding_methods.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/decoding_methods.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/decoding_methods.inc
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/protocol_constants.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/Serializer.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/Serializer.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/utils.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ffi/ir_stream/utils.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ir/parsing.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ir/parsing.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ir/parsing.inc
    ${clp_SOURCE_DIR}/components/core/src/clp/ir/types.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ReaderInterface.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/ReaderInterface.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/string_utils/string_utils.cpp
    ${clp_SOURCE_DIR}/components/core/src/clp/string_utils/string_utils.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/TraceableException.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/time_types.hpp
    ${clp_SOURCE_DIR}/components/core/src/clp/type_utils.hpp
    src/ffi_go/types.hpp
    src/ffi_go/ir/deserializer.cpp
    src/ffi_go/ir/types.hpp
    src/ffi_go/ir/serializer.cpp
    src/lint/msgpack.hpp
)

install(TARGETS ${LIB_NAME}
    FILE_SET HEADERS
)
