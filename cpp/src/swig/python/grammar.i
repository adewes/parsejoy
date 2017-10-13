%include <std_map.i>
%include <std_vector.i>
%include <std_shared_ptr.i>

%{
#include "grammar.h"
static PyObject* pMyException;  /* add this! */
%}

%init %{
    pMyException = PyErr_NewException("parsejoy.Exception", NULL, NULL);
    Py_INCREF(pMyException);
    PyModule_AddObject(m, "MyException", pMyException);
%}

%exception {
    try {
        $action
    } catch (std::exception &e) {
        PyErr_SetString(pMyException, const_cast<char*>(e.what()));
        SWIG_fail;
    }
}


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

%template(RuleVector) vector<std::shared_ptr<Rule>>;

shared_ptr<MapRule> toMapRule(shared_ptr<Rule> rule);
MapRule& toMapRule(Rule& rule);
SequenceRule& toSequenceRule(Rule& rule);
StringRule& toStringRule(Rule& rule);
