

/*we need this to map UserData (int) types to real variables...*/
namespace sscientists{
namespace parsejoy{

//we store the userdata into the registry
%typemap(in) (UserData) {
	$1 = (int) luaL_ref(L, LUA_REGISTRYINDEX);
}

//we return the userdata object from the registry
%typemap(out) (UserData) {
	lua_rawgeti(L, LUA_REGISTRYINDEX, $1);
	SWIG_arg++;
}

}
}

%{
#include "parser.h"
%}

%catches(std::runtime_error) sscientists::parsejoy::ParserGenerator::compile();
%catches(std::runtime_error) sscientists::parsejoy::ParserGenerator::compile(const std::string& str);

%header{
namespace sscientists{namespace parsejoy{

ParserResult runParser(Parser parser, State& state){
    return parser(state);
};

bool getError(ParserResult result){
    return std::get<1>(result);
}

shared_ptr<Token> getToken(ParserResult result){
    return std::get<0>(result);
}

}}};

namespace sscientists{
namespace parsejoy{

class State {
private:
    State(LuaEnvironment& luaEnvironment);
    State(const State& obj);
public:
    UserData getUserData();
    void setUserData(UserData userData);
    virtual const std::string className(){return "sscientists::parsejoy::State";};
};

class ParserGenerator {
private:
    ParserGenerator(const shared_ptr<Rule> grammar, LuaEnvironment& environment);
public:
    Parser compile();
    bool debug;
};

ParserResult runParser(Parser parser, State& state);
shared_ptr<Token> getToken(ParserResult result);
bool getError(ParserResult result);

}}
