#include "stringparser.h"
#include <iostream>
#include <regex>

using namespace std;

namespace sscientists{
namespace parsejoy{

StringState::StringState(LuaEnvironment& env,const std::string s) : State(env) {
    s_ = s;
    pos_ = 0;
}


StringState::StringState(const StringState& state) : State(state) {
    s_ = state.s_;
    pos_ = state.pos_;
}


std::string StringState::Value(const unsigned int n){
    return s_.substr(pos_,n);
}

std::string StringState::Value(){
    return s_.substr(pos_);
}

std::string::const_iterator StringState::Current(){
    return s_.begin()+pos_;
}

std::string::const_iterator StringState::End(){
    return s_.end();
}

unsigned int StringState::Size(){
    return s_.size();
}

void StringToken::SetNext(std::shared_ptr<Token> token){

}

StringToken::StringToken(const std::string type, const std::string value,const Position& pos) : Token() {
    pos_ = pos;
    value_ = value;
    type_ = type;
}

void StringParserGenerator::setPrefixes(std::string name,Prefix& prefix) {

};

Parser StringParserGenerator::compileLiteral(const shared_ptr<Rule> rule, RuleCache& cache){
    auto stringRule = dynamic_pointer_cast<StringRule>(rule);
    if (stringRule == nullptr){
        auto subRule = extractSubrule(rule, "$literal");
        stringRule = dynamic_pointer_cast<StringRule>(subRule);
        if (stringRule == nullptr)
            throw std::runtime_error("Expected either a dictionary or a string!");
    }
    auto literal = stringRule->rule;
    auto literalParser = [literal, this](State& state){
        if (debug)
            cout << "Parsing literal " << literal << "\n";
        StringState& stringState = static_cast<StringState&>(state);
        auto value = stringState.Value(literal.size());
        if (value != literal)
            return ParserResult(emptyToken, true);
        if (debug)
            cout << "Success!\n";
        Position position{from:stringState.Position(),to:stringState.Position()+1};
        auto token = make_shared<StringToken>(literal, value, position);
        stringState.Advance(literal.size());
        return ParserResult(std::static_pointer_cast<Token>(token), false);
    };

    return literalParser;
}

Parser StringParserGenerator::compileEof(const shared_ptr<Rule> rule, RuleCache& cache){
    bool debug = this->debug;
    auto eofParser = [debug](State& state){
        if (debug)
            cout << "Parsing EOF\n";
        StringState& stringState = static_cast<StringState&>(state);
        if (stringState.Position() == stringState.Size()){
            return ParserResult(emptyToken, false);
        }
        return ParserResult(emptyToken, true);
    };

    return eofParser;
}

Parser StringParserGenerator::compileRegex(const shared_ptr<Rule> rule, RuleCache& cache){

    auto subRule = extractSubrule(rule, "$regex");

    std::string regexString;
    try{
        auto regexRule = dynamic_pointer_cast<StringRule>(subRule);
        regexString = regexRule->rule;
    } catch (const std::bad_cast& e){
        try{
            auto regexRule = dynamic_pointer_cast<MapRule>(subRule);
            auto regexStringEntry = regexRule->rules.find("regex");
            if (regexStringEntry == regexRule->rules.end())
                throw std::runtime_error("Missing 'regex' entry!");
            auto stringRule = dynamic_pointer_cast<StringRule>(regexStringEntry->second);
            regexString = stringRule->rule;
        } catch (const std::bad_cast& e){
            throw std::runtime_error("Expected a string or map rule!");
        }
    }
    bool debug = debug;
    auto baseRegex = std::regex("^"+regexString);
    auto regexParser = [debug, baseRegex, regexString](State& state){
        StringState& stringState = static_cast<StringState&>(state);
        std::smatch baseMatch;
        if (debug)
            cout << "Parsing regex " << regexString << "\n";
        if (std::regex_search(stringState.Current(), stringState.End(), baseMatch, baseRegex)){
            if (debug)
                cout << "Success!\n";
            auto fullMatch = baseMatch[0].str();
            Position position{from:stringState.Position(),to:stringState.Position()+(unsigned int)fullMatch.size()};
            auto token = make_shared<StringToken>("regex", fullMatch, position);
            stringState.Advance(fullMatch.size());
            return ParserResult(std::static_pointer_cast<Token>(token), false);
        } else {
            return ParserResult(emptyToken, true);
        }
    };

    return regexParser;
}

ruleParser StringParserGenerator::resolveRule(const std::string name){
    try{
        return ParserGenerator::resolveRule(name);
    } catch(const std::runtime_error& e) {
        if (name == "$eof")
            return std::bind(&StringParserGenerator::compileEof, this, placeholders::_1, placeholders::_2);
        if (name == "$literal")
            return std::bind(&StringParserGenerator::compileLiteral, this, placeholders::_1, placeholders::_2);
        if (name == "$regex")
            return std::bind(&StringParserGenerator::compileRegex, this, placeholders::_1, placeholders::_2);
        //if nothing else matches, we assume this is a literal
        return std::bind(&StringParserGenerator::compileLiteral, this, placeholders::_1, placeholders::_2);
    }
}

Parser StringParserGenerator::wrap(Parser parser, std::string name, const shared_ptr<Rule> rule,bool emit){
    auto wrapper = [name, parser, this](State& state){
        auto result = parser(state);
        return result;
    };
    return wrapper;
}

}
}
