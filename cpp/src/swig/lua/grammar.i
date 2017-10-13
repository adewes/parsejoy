%include "shared_ptr.i"
%include <std_map.i>

%{
#include "grammar.h"
%}

%template(RuleMap) std::map<std::string, std::shared_ptr<sscientists::parsejoy::Rule>>;

%include "grammar.h"

%header{

namespace sscientists{
namespace parsejoy{

MapRule& toMapRule(Rule& rule){
    return dynamic_cast<MapRule&>(rule);
}

SequenceRule& toSequenceRule(Rule& rule){
    return dynamic_cast<SequenceRule&>(rule);
}

StringRule& toStringRule(Rule& rule){
    return dynamic_cast<StringRule&>(rule);
}

shared_ptr<MapRule> toMapRule(shared_ptr<Rule> rule){
    return dynamic_pointer_cast<MapRule>(rule);
}

}}};

using namespace sscientists::parsejoy;

wrap_shared_ptr(RuleSharedPtr, Rule);
wrap_shared_ptr(MapRuleSharedPtr, MapRule);
wrap_shared_ptr(StringRuleSharedPtr, StringRule);

//%template(SharedPtrRule) shared_ptr<Rule>;
//%template(SharedPtrStringRule) shared_ptr<StringRule>;
//%template(SharedPtrMapRule) shared_ptr<MapRule>;

shared_ptr<MapRule> toMapRule(shared_ptr<Rule> rule);
MapRule& toMapRule(Rule& rule);
SequenceRule& toSequenceRule(Rule& rule);
StringRule& toStringRule(Rule& rule);
