%module parsejoy

%include <std_string.i>
%include <std_vector.i>
%include <std_except.i>

%typemap(throws) std::runtime_error
%{
  lua_pushstring(L,$1.what()); // assuming my_except::what() returns a const char* message
  SWIG_fail; // trigger the error handler
%}

%include "lua_environment.i"
%include "grammar.i"
%include "parser.i"
%include "stringparser.i"
%include "set.i"
