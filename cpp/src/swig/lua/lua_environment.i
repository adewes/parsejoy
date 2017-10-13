%{
#include "lua_environment.h"
%}

%include "lua_environment.h"

%catches(std::exception) sscientists::parsejoy::loadYAML();
%catches(std::exception) sscientists::parsejoy::parseYAMLGrammar();
