#pragma once

#include <string>
#include <map>
#include <vector>
#include <memory>
#include "yaml-cpp/yaml.h"

using namespace std;

namespace sscientists {
namespace parsejoy {

class GrammarException : public exception {
public:
  string description;
  GrammarException(string s);
};

enum Type {SEQUENCE, MAP, STRING};

class Rule {
public:
    virtual Type type() const = 0;
};

/*!
  A sequence rule, i.e. [foo, bar, baz]
*/
class SequenceRule : public Rule {
public:
    std::vector<std::shared_ptr<Rule>> rules;
    Type type() const;
};

/*!
  A map rule, i.e. {test : foo, bar: baz}
*/
class MapRule : public Rule {
public:
    std::map<std::string,std::shared_ptr<Rule>> rules;
    Type type() const;
};

/*!
  A string rule, i.e. "foo"
*/
class StringRule : public Rule {
public:
    StringRule(std::string str);
    Type type() const;
    std::string rule;
};

const YAML::Node loadYAML(const std::string filename);
shared_ptr<Rule> parseYAMLGrammar(const YAML::Node &grammar);

}
}
