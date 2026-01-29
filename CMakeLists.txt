cmake_minimum_required(VERSION 3.10)
project(chippy)

set(CMAKE_CXX_STANDARD 17)

file(GLOB_RECURSE PROJECT_SOURCES "src/*.cpp")

add_executable(${PROJECT_NAME} ${PROJECT_SOURCES})

target_include_directories(${PROJECT_NAME} PRIVATE
    ${PROJECT_SOURCE_DIR}/src
)

if(MSVC)
    target_compile_options(${PROJECT_NAME} PRIVATE /W4)
else()
    target_compile_options(${PROJECT_NAME} PRIVATE -Wall -Wextra -Wpedantic)
endif()