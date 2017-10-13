#pragma once

#include "parser.h"

namespace sscientists{
namespace parsejoy{

struct Position {
    unsigned int from;
    unsigned int to;
};

class StringSavepoint : public Savepoint {
public:
    unsigned int pos_;
};

class StringState : public State {
public:
    StringState(const StringState& state);
    StringState(LuaEnvironment& env, const std::string s);
    ~StringState() override {};
    std::string Value(const unsigned int n);
    std::string::const_iterator Current();
    std::string::const_iterator End();
    std::string Value();
    std::unique_ptr<Savepoint> Save() override;
    std::unique_ptr<Savepoint> Save(std::unique_ptr<StringSavepoint>);
    void Restore(std::unique_ptr<Savepoint>) override;
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
private:
    Position pos_;
    std::string value_;
    std::string type_;
};

class StringParserGenerator : public ParserGenerator {
public:
    StringParserGenerator(const shared_ptr<Rule> rule, LuaEnvironment& env) : ParserGenerator(rule, env) {};
    void setPrefixes(std::string name,Prefix& prefix);
    //Set getPrefix(const shared_ptr<Rule> rule,void *args);
    //Set getEmptyPrefix(void *args);
    Parser compileLiteral(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileEof(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileRegex(const shared_ptr<Rule> rule, RuleCache& cache);
    ruleParser resolveRule(const std::string name) override;
    Parser wrap(Parser parser, std::string name, const shared_ptr<Rule> rule,bool emit) override;

};

}
}
