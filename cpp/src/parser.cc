#include "grammar.h"
#include "parser.h"
#include "sha.h"
#include <iostream>
#include <string>

using namespace std;

std::string string_to_hex(const std::string& input)
{
    static const char* const lut = "0123456789ABCDEF";
    size_t len = input.length();

    std::string output;
    output.reserve(2 * len);
    for (size_t i = 0; i < len; ++i)
    {
        const unsigned char c = input[i];
        output.push_back(lut[c >> 4]);
        output.push_back(lut[c & 15]);
    }
    return output;
}

namespace sscientists {
namespace parsejoy {

State::~State() {
}

State::State(LuaEnvironment& luaEnvironment) : luaEnvironment_(luaEnvironment) {
}

State::State(const State& state) : luaEnvironment_(state.luaEnvironment_) {
}

unsigned int State::Advance(const unsigned int n){
    pos_ += n;
    return pos_;
}

unsigned int State::Position(){
    return pos_;
}

void State::SetPosition(unsigned int n){
    pos_ = n;
}

ParserGenerator::ParserGenerator(const shared_ptr<Rule> g, LuaEnvironment& env) : luaEnvironment_(env) {
    debug = false;
    assert(g->type() == Type::MAP);
    grammar = static_pointer_cast<MapRule>(g);
};

std::string ParserGenerator::getFingerprint(const shared_ptr<Rule> rule){

    CryptoPP::SHA256 sha256;

    const std::string value = "foo";

    sha256.Update((const byte *)value.c_str(),value.size());

    switch(rule->type()){
        case Type::STRING:{
            auto stringRule = static_pointer_cast<StringRule>(rule);
            sha256.Update((const byte *)stringRule->rule.c_str(),stringRule->rule.size());
            break;
        }
        case Type::SEQUENCE:{
            auto sequenceRule = static_pointer_cast<SequenceRule>(rule);
            for(auto it=sequenceRule->rules.begin();it!=sequenceRule->rules.end();++it){
                std::string result = getFingerprint(*it);
                sha256.Update((const byte *)result.c_str(),result.size());
            }
            break;
        }
        case Type::MAP:{
            auto mapRule = static_pointer_cast<MapRule>(rule);
            for(auto it=mapRule->rules.begin();it!=mapRule->rules.end();++it){
                std::string key = it->first;
                sha256.Update((const byte *)key.c_str(),key.size());
                std::string result = getFingerprint(it->second);
                sha256.Update((const byte *)result.c_str(),result.size());
            }
            break;
        }
        default:{
            break;
        }
    }

    byte* bytes = new byte[sha256.DigestSize()];
    sha256.Final(bytes);
    auto s = reinterpret_cast<const char *>(bytes);
    return std::string(s);
}

//Set& ParserGenerator::getPrefixes(const shared_ptr<Rule> rule,HashSet& visitedRules,void *args,bool onlyStart){
    /*
    stringRule, ok := rule.(string)

    if ok {
        if visitedRules.Contains(stringRule) {
            return pg.plugin.getEmptyPrefix(args)
        }
        visitedRules.Add(stringRule)
        grammarRule, ok := pg.grammar[stringRule]
        if ok {
            prefixes := pg.getPrefixes(grammarRule,visitedRules,args, onlyStart)
            pg.plugin.setPrefixes(stringRule, prefixes)
            return prefixes
        }
        return pg.plugin.getPrefix(stringRule,args)
    }

    listRule, ok := rule.([]interface{})

    if ok && len(listRule) > 0{
        prefixes := pg.plugin.getEmptyPrefix(args)
        for i:= range listRule  {
            ruleItem := listRule[i]
            subrulePrefixes := pg.getPrefixes(ruleItem, visitedRules,args,onlyStart)
            //if the current prefixes contain a false value (i.e. an optional token), we first remove it
            //before adding the new prefixes
            if prefixes.Contains(false) {
                _ = prefixes.Remove(false)
            }
            prefixes, _ = prefixes.Union(subrulePrefixes)
            //if we match only the start and this contains no false value, we break from the loop
            if onlyStart && !prefixes.Contains(false){
                break
            }
        }
        return prefixes
    }

    dictRule, ok := rule.(map[interface{}]interface{})

    if ok {
        if len(dictRule) == 1 {
            var key, value interface{}
            for key, value = range dictRule {
                break
            }
            stringKey, ok := key.(string)
            if !ok {
                panic("This should not happen!")
            }
            switch stringKey {
                case "$or":
                    listValue, ok := value.([]interface{})
                    if !ok || len(listValue) == 0 {
                        panic("This should not happen!")
                    }
                    prefixes := pg.plugin.getEmptyPrefix(args)
                    for i := range listValue {
                        subrule := listValue[i]
                        newPrefixes := pg.getPrefixes(subrule, visitedRules, args, onlyStart)
                        prefixes, _ = prefixes.Union(newPrefixes)
                    }
                    return prefixes
                case "$ast-prop": fallthrough
                case "$ast-node":
                    dictValue, ok := value.(map[interface{}]interface{})
                    if !ok {
                    }
                    return pg.getPrefixes(dictValue["value"], visitedRules, args, onlyStart)
                case "$and": fallthrough
                case "$repeat":
                    return pg.getPrefixes(value, visitedRules, args, onlyStart)
                case "$not":
                    nilPrefix := pg.plugin.getEmptyPrefix(args)
                    nilPrefix.Add(nil)
                    return nilPrefix
                case "$optional":
                    newPrefixes := pg.getPrefixes(value, visitedRules,args, onlyStart)
                    newPrefixes.Add(false)
                    return newPrefixes
                default:
                    return pg.plugin.getPrefix(rule,args)
            }
        }
    }
    panic("this should never happen")
    */
//}

Parser ParserGenerator::compileLua(const shared_ptr<Rule> rule, RuleCache& cache){

    const shared_ptr<Rule> subRule = extractSubrule(rule, "$lua");
    auto stringRule = dynamic_pointer_cast<StringRule>(subRule);

    auto compiledLua = luaEnvironment_.compileCode(stringRule->rule);

    auto luaParser = [compiledLua, this](State& state){
        unsigned int pos = state.Position();
        //we place the state as a global variable
        //we can actually do this only once while the parser is running...
        this->luaEnvironment_.place("state", (void *)&state,state.className(), false);
        auto result = this->luaEnvironment_.runByteCode(compiledLua);
        auto error = result != 0;
        if (error)
            state.SetPosition(pos);
        return ParserResult(emptyToken, error);
    };

    return luaParser;

}


Parser ParserGenerator::compileSequence(const shared_ptr<Rule> rule, RuleCache& cache){

    std::vector<Parser> parsers;
    auto sequenceRule = dynamic_pointer_cast<SequenceRule>(rule);

    for(auto it=sequenceRule->rules.begin();it!=sequenceRule->rules.end();++it){
        auto parser = wrap(compileRule(*it, "", cache),"$sequence", rule, false);
        parsers.push_back(parser);
    }

    auto sequenceParser = [parsers](State& state){
        unsigned int pos = state.Position();
        for(auto it=parsers.begin();it!=parsers.end();++it){
            auto result = (*it)(state);
            bool error = std::get<1>(result);
            if (error == true){
                state.SetPosition(pos);
                return ParserResult(emptyToken, true);
            }
        }
        return ParserResult(emptyToken, false);
    };

    return sequenceParser;

}

Parser ParserGenerator::wrap(Parser parser, std::string name, const shared_ptr<Rule> rule, bool emit){
    auto wrapper = [name, parser, this](State& state){
        cout << &state << "\n";
        auto result = parser(state);
        return result;
    };
    return wrapper;
}

const shared_ptr<Rule> ParserGenerator::extractSubrule(const shared_ptr<Rule> rule, const std::string name){
    /*
    We might not need this function anymore...
    */
    try{
        auto mapRule = dynamic_pointer_cast<MapRule>(rule);
        auto it = mapRule->rules.find(name);
        if (it == mapRule->rules.end())
            throw std::runtime_error("Expected a "+name+" entry!");
        return it->second;
    } catch (const std::bad_cast& e){
        throw std::runtime_error("Expected a map rule in extractSubrule!");
    }
}

Parser ParserGenerator::compileRepeat(const shared_ptr<Rule> rule, RuleCache& cache){

    const shared_ptr<Rule> subRule = extractSubrule(rule, "$repeat");
    auto subruleParser = compileRule(subRule, "", cache);

    auto repeatParser = [subruleParser](State& state){
        bool success = false;
        unsigned int pos = state.Position();
        while(true){
            auto result = subruleParser(state);
            bool error = std::get<1>(result);
            if (error){
                if (success)
                    return ParserResult(emptyToken, false);
                state.SetPosition(pos);
                return result;
            }
            success = true;
        }
    };

    return repeatParser;
}

Parser ParserGenerator::compileOr(const shared_ptr<Rule> rule,RuleCache& cache){

    const shared_ptr<Rule> subRule = extractSubrule(rule, "$or");
    std::vector<Parser> parsers;
    auto sequenceRule = dynamic_pointer_cast<SequenceRule>(subRule);

    for(auto it=sequenceRule->rules.begin();it!=sequenceRule->rules.end();++it){
        auto parser = compileRule(*it, "", cache);
        parsers.push_back(parser);
    }

    auto orParser = [this,parsers](State& state){
        for(auto it=parsers.begin();it!=parsers.end();++it){
            unsigned int pos = state.Position();
            auto result = (*it)(state);
            bool error = std::get<1>(result);
            if (!error)
                return result;
            state.SetPosition(pos);
            }
        return ParserResult(emptyToken, true);
    };

    return orParser;

}

Parser ParserGenerator::compileAnd(const shared_ptr<Rule> rule, RuleCache& cache){

    const shared_ptr<Rule> subRule = extractSubrule(rule, "$and");
    auto subruleParser = compileRule(subRule, "", cache);

    auto andParser = [subruleParser](State& state){
        unsigned int pos = state.Position();
        auto result = subruleParser(state);
        state.SetPosition(pos);
        bool error = std::get<1>(result);
        if (error)
            return ParserResult(emptyToken, true);
        return ParserResult(emptyToken, false);
    };

    return andParser;
}

Parser ParserGenerator::compileNot(const shared_ptr<Rule> rule, RuleCache& cache){

    const shared_ptr<Rule> subRule = extractSubrule(rule, "$not");
    auto subruleParser = compileRule(subRule, "", cache);

    auto notParser = [subruleParser](State& state){
        unsigned int pos = state.Position();
        auto result = subruleParser(state);
        state.SetPosition(pos);
        bool error = std::get<1>(result);
        if (error)
            return ParserResult(emptyToken, false);
        return ParserResult(emptyToken, true);
    };

    return notParser;
}

Parser ParserGenerator::compileOptional(const shared_ptr<Rule> rule, RuleCache& cache){

    const shared_ptr<Rule> subRule = extractSubrule(rule, "$optional");
    auto subruleParser = compileRule(subRule, "", cache);

    auto optionalParser = [subruleParser](State& state){
        unsigned int pos = state.Position();
        auto result = subruleParser(state);
        bool error = std::get<1>(result);
        if (error){
            state.SetPosition(pos);
            return ParserResult(emptyToken, false);
        }
        return result;
    };
    return optionalParser;
}

ruleParser ParserGenerator::resolveRule(const std::string name){
    cout << "Resolving rule |" << name << "|\n";
    if (name == "$lua")
        return std::bind(&ParserGenerator::compileLua, this, placeholders::_1, placeholders::_2);
    if (name == "$repeat")
        return std::bind(&ParserGenerator::compileRepeat, this, placeholders::_1, placeholders::_2);
    if (name == "$or")
        return std::bind(&ParserGenerator::compileOr, this, placeholders::_1, placeholders::_2);
    if (name == "$and")
        return std::bind(&ParserGenerator::compileAnd, this, placeholders::_1, placeholders::_2);
    if (name == "$not")
        return std::bind(&ParserGenerator::compileNot, this, placeholders::_1, placeholders::_2);
    if (name == "$optional")
        return std::bind(&ParserGenerator::compileOptional, this, placeholders::_1, placeholders::_2);
    throw std::runtime_error("Could not resolve rule "+name+"!");
}

Parser ParserGenerator::compileRule(const shared_ptr<Rule> rule, const std::string name, RuleCache& cache){

    /*
        Compiles a given rule.
    */

    if (name != "") {
        //we check if we have the rule in the cache. If yes, we return it from there...
        if (cache.find(name) != cache.end())
            return cache.find(name)->second;
        auto subruleParser = [name, &cache](State& state) {
            if (cache.find(name) == cache.end())
                throw std::runtime_error("Unable to find parser for rule "+name);
            return cache[name](state);
        };
        if (debug)
            cout << "Setting cache function for name " << name << "\n";
        cache[name] = subruleParser;
    }

    auto parseSubrule = [&](const shared_ptr<MapRule> rule, std::string subruleName){
        if (rule->rules.size() == 1){
            std::string ruleName = rule->rules.begin()->first;
            const shared_ptr<Rule> subRule = rule->rules.begin()->second;
            auto generator = resolveRule(ruleName);
            try{
                return wrap(generator(rule, cache), ruleName, rule, false);
            }catch(const std::runtime_error& e){
                cerr << "Error when parsing rule " << ruleName << " within context of " << name << "\n";
                throw;
            }

        } else {
            throw std::runtime_error("expected exactly one rule in subrule "+subruleName+" when parsing rule "+name);
        }
    };

    Parser parser;

    switch (rule->type()){
        case Type::STRING:{
            /*
            This is just a string, referencing another rule from the grammar.
            */

            auto newName = static_pointer_cast<StringRule>(rule)->rule;
            auto it = parserMap.find(name);
            if (it != parserMap.end()){
                parser = it->second;
            } else {
                auto grammarRule = grammar->rules.find(newName);
                if (grammarRule != grammar->rules.end()){
                    parser = compileRule(grammarRule->second, newName, cache);
                } else {
                    auto generator = resolveRule(newName);
                    parser = generator(rule, cache);
                }
            }
            break;
        }
        case Type::MAP:{
            auto mapRule = static_pointer_cast<MapRule>(rule);
            parser = parseSubrule(mapRule, name);
            break;
        }
        case Type::SEQUENCE:{
            parser = wrap(compileSequence(rule, cache), name, rule, true);
            break;
        }
    }

    if (name != ""){
        cache[name] = parser;
    }

    return parser;

}

Parser ParserGenerator::compile(){
    return compile("start");
}

Parser ParserGenerator::compile(const std::string& startRule){
    auto parser = compileRule(grammar->rules[startRule],startRule, cache);
    return parser;
}


}
}
