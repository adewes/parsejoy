%{
#include "stringparser.h"
%}

#include "parser.h"

%import "parser.i"

namespace sscientists{
namespace parsejoy{

struct Position {
    unsigned int from;
    unsigned int to;
};

class StringState : public State {
public:
    StringState(const StringState& state);
    StringState(LuaEnvironment& env, const std::string s);
    std::string Value(const unsigned int n);
    std::string::const_iterator Current();
    std::string::const_iterator End();
    std::string Value();
    const std::string className() override {return "sscientists::parsejoy::StringState";};
    unsigned int Advance(const unsigned int n);
    unsigned int Position();
    unsigned int Size();
private:
    std::string s_;
    unsigned int pos_;
};

class StringToken : public Token {
public:
    StringToken(const std::string type_, const std::string value, const Position& pos);
    void SetNext(std::shared_ptr<Token> token) override;
};

class StringParserGenerator : public ParserGenerator {
public:
    StringParserGenerator(const shared_ptr<Rule> rule, LuaEnvironment& env) : ParserGenerator(rule, env) {};
};

}
}