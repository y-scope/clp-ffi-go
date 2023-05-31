cmake_minimum_required(VERSION 3.5.1)

project(clp_ffi
    LANGUAGES CXX C
)

# Set default build type
if (NOT CMAKE_BUILD_TYPE AND NOT CMAKE_CONFIGURATION_TYPES)
    set(default_build_type "Release")
    message(STATUS "No build type specified. Setting to '${default_build_type}'.")
    set(CMAKE_BUILD_TYPE "${default_build_type}" CACHE STRING "Choose the type of build." FORCE)
endif()

option(BUILD_SHARED_LIBS "Build using shared libraries" ON)

set(LIB_NAME "clp_ffi" CACHE STRING "Library name containing os and arch.")
if (DEFINED ENV{GOOS})
    string(APPEND LIB_NAME "_$ENV{GOOS}")
endif()
if (DEFINED ENV{GOARCH})
    string(APPEND LIB_NAME "_$ENV{GOARCH}")
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
    PRIVATE
    cxx_std_17
)

target_include_directories(${LIB_NAME}
    PRIVATE
    ${CMAKE_SOURCE_DIR}
    ${CMAKE_SOURCE_DIR}/src
)

# target_compile_options(${LIB_NAME}
#     PRIVATE
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

install(TARGETS ${LIB_NAME})
