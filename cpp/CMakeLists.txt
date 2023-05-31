cmake_minimum_required(VERSION 3.5.1)

project(clp_ffi
    VERSION 0.0.1
    LANGUAGES CXX C
)

# Set default build type
if (NOT CMAKE_BUILD_TYPE AND NOT CMAKE_CONFIGURATION_TYPES)
    set(default_build_type "Release")
    message(STATUS "No build type specified. Setting to '${default_build_type}'.")
    set(CMAKE_BUILD_TYPE "${default_build_type}" CACHE STRING "Choose the type of build." FORCE)
endif()

option(BUILD_SHARED_LIBS "Build using shared libraries" ON)

if (NOT DEFINED GOOS)
    if (DEFINED ENV{GOOS})
        set(GOOS "$ENV{GOOS}" CACHE STRING "GOOS name (see `go env`)" FORCE)
    else()
        message(WARNING "GOOS not specified. Setting to empty string.")
        set(GOOS "")
    endif()
endif()

if (NOT DEFINED GOARCH)
    if (DEFINED ENV{GOARCH})
        set(GOARCH "$ENV{GOARCH}" CACHE STRING "GOARCH name (see `go env`)" FORCE)
    else()
        message(WARNING "GOARCH not specified. Setting to empty string.")
        set(GOARCH "")
    endif()
endif()

set(LIB_NAME "clp_ffi_${GOOS}_${GOARCH}" CACHE STRING "Full library name containing os and arch.")
if (BUILD_SHARED_LIBS)
    set(LIB_FILE "${CMAKE_SHARED_LIBRARY_PREFIX}${LIB_NAME}${CMAKE_SHARED_LIBRARY_SUFFIX}" CACHE STRING "Full library file name containing os, arch, and file prefix and suffix.")
else()
    set(LIB_FILE "${CMAKE_STATIC_LIBRARY_PREFIX}${LIB_NAME}${CMAKE_STATIC_LIBRARY_SUFFIX}" CACHE STRING "Full library file name containing os, arch, and file prefix and suffix.")
endif()

add_library(${LIB_NAME}
    clp/components/core/src/Defs.h
    clp/components/core/src/ffi/ir_stream/encoding_methods.cpp
    clp/components/core/src/ffi/ir_stream/encoding_methods.hpp
    clp/components/core/src/ffi/ir_stream/decoding_methods.cpp
    clp/components/core/src/ffi/ir_stream/decoding_methods.hpp
    clp/components/core/src/ffi/encoding_methods.cpp
    clp/components/core/src/ffi/encoding_methods.hpp
    clp/components/core/src/ffi/encoding_methods.tpp
    src/log_event.h
    src/LogEvent.cpp
    src/LogEvent.hpp
    src/ir/encoding.cpp
    src/ir/encoding.h
    src/ir/decoding.cpp
    src/ir/decoding.h
    src/message/encoding.cpp
    src/message/encoding.h
)

target_compile_features(${LIB_NAME}
    PRIVATE cxx_std_17
)

target_include_directories(${LIB_NAME}
    PRIVATE
    ${CMAKE_SOURCE_DIR}
    ${CMAKE_SOURCE_DIR}/src
)

# cries in clpain
# target_compile_options(${LIB_NAME} PRIVATE
#     $<$<CXX_COMPILER_ID:MSVC>:/W4 /WX>
#     $<$<NOT:$<CXX_COMPILER_ID:MSVC>>:-Wall -Wextra -Wpedantic -Werror>
#     )

# Macro providing the length of the absolute source directory path so we can
# create a relative (rather than absolute) __FILE__ macro
string(LENGTH "${CMAKE_SOURCE_DIR}/" SOURCE_PATH_SIZE)
target_compile_definitions(${LIB_NAME}
    PUBLIC
    SOURCE_PATH_SIZE=${SOURCE_PATH_SIZE}
)

install(
    FILES
    ${CMAKE_CURRENT_BINARY_DIR}/${LIB_FILE}
    DESTINATION
    .
)
