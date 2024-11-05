#pragma once

#include <tuple>
#include <map>
#include "grammar.h"
#include "lua_environment.h"
#include "set.h"

namespace sscientists {
namespace parsejoy {

typedef unsigned char byte;

class State {
public:
    State(LuaEnvironment& luaEnvironment);
    State(const State& obj);
    unsigned int Advance(const unsigned int n);
    unsigned int Position();
    void SetPosition(unsigned int n);
    virtual ~State() = 0;
    virtual const std::string className(){return "sscientists::parsejoy::State";};
protected:
    unsigned int pos_;
private:
    LuaEnvironment& luaEnvironment_;
};

class Token {
public:
    virtual void SetNext(std::shared_ptr<Token> token) = 0;
};

static std::shared_ptr<Token> emptyToken;

class Prefix {

};

class CompilationError : public std::exception {

};

class ParsingError : public std::exception {

};


typedef std::tuple<std::shared_ptr<Token>, bool> ParserResult;
typedef std::function<ParserResult(State&)> Parser;
typedef std::map<std::string,Parser> RuleCache;
typedef std::function<Parser(const shared_ptr<Rule>, RuleCache& cache)> ruleParser;

class ParserGenerator {
public:
    ParserGenerator(const shared_ptr<Rule> grammar, LuaEnvironment& environment);
    std::string getFingerprint(const shared_ptr<Rule> rule);
    //Set& getPrefixes(const shared_ptr<Rule> rule,HashSet& visitedRules,void *args,bool onlyStart);
    const shared_ptr<Rule> extractSubrule(const shared_ptr<Rule> rule, const std::string name);
    Parser compileLua(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileSequence(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileRepeat(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileOr(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileAnd(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileNot(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileOptional(const shared_ptr<Rule> rule, RuleCache& cache);
    Parser compileRule(const shared_ptr<Rule> rule, const std::string name, RuleCache& cache);
    Parser compile(const std::string& startRule);
    Parser compile();

    virtual Parser wrap(Parser parser, std::string name, const shared_ptr<Rule> rule, bool emit);
    virtual ruleParser resolveRule(const std::string name);
    virtual void setPrefixes(std::string name,Prefix& prefix) = 0;
    //virtual Set getPrefix(const shared_ptr<Rule> rule,void *args) = 0;
    //virtual Set getEmptyPrefix(void *args) = 0;
    bool debug;
private:
    shared_ptr<MapRule> grammar;
    RuleCache cache;
    LuaEnvironment& luaEnvironment_;
    std::map<std::string,Parser> parserMap;
};

}
}
