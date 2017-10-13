#include "grammar.h"
#include "yaml-cpp/yaml.h"
#include <iostream>

using namespace std;

namespace sscientists {
namespace parsejoy{

GrammarException::GrammarException(string s){
  description = s;
}

const YAML::Node loadYAML(const std::string filename){
    auto node = YAML::LoadFile(filename);
    return node;
}

shared_ptr<Rule> parseYAMLGrammar(const YAML::Node &grammar){
  YAML::const_iterator it;
  switch (grammar.Type()){
    case YAML::NodeType::Scalar:
      return make_shared<StringRule>(grammar.as<string>());
    case YAML::NodeType::Sequence:
    {
      auto sequenceRule = make_shared<SequenceRule>();
      for(it=grammar.begin();it!=grammar.end();++it){
        sequenceRule->rules.push_back(parseYAMLGrammar(*it));
      }
      return sequenceRule;
    }
    case YAML::NodeType::Map:
    {
      auto mapRule = make_shared<MapRule>();
      for(it=grammar.begin();it!=grammar.end();++it){
        mapRule->rules[it->first.as<string>()] = parseYAMLGrammar(it->second);
      }
      return mapRule;
    }
  };
  throw GrammarException("Unknown node type!");
}

Type SequenceRule::type() const{
    return Type::SEQUENCE;
}

Type MapRule::type() const{
    return Type::MAP;
}

Type StringRule::type() const{
    return Type::STRING;
}

StringRule::StringRule(std::string str){
  rule = str;
}

}
}
