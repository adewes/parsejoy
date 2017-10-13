# Parsejoy - C++ Version

This folder contains the C++ source code of the Parsejoy parser generator. It
provides several things:

## The parsejoy command line tool

The parsejoy command line tool is the easiest way to get started with building
a parser. It contains an interactive Lua interpreter that provides bindings
to all relevant C++ classes as well as helper functions and allows you to
control the parser via Lua.

## A Python binding (Python 2.7 / Python 3.5)

The Python binding exposes the Parsejoy functionality to Python.

## Build Instructions

### Source Code Dependencies

The standalone library requires the following dependencies:

* Boost
* Luajit
* YAML-CPP

The Python binding requires a Python development environment.

The command line tool requires `libreadline` (though it is possible to compile
it without support for this).

### Compiling

To build the tool:

    make tool

To build the Python 2 library:

    make python2lib

To build the Python 3 library:

    make python3lib

Please note: If you build both Python2/Python3 versions make sure to run
`make clean` between the two runs, as the bindings that SWIG generate are
slightly different for the two versions and will not be rebuilt automatically.

## Getting Started

To test the tool, you can run one of the example scripts in the `scripts`
folder:

    ./build/tool -f scripts/lua/calculator_example.lua

This will compile the simple calculator grammar and use it to parse an example
input.