%Module(name=parsejoy.parser_generator,version=1,call_super_init=False)

class ParserPlugin /Abstract/{
%TypeHeaderCode
#include "parser_generator.h"
using namespace sscientists::parsejoy;
%End
public:
    ParserPlugin();
};

class ParserGenerator {

%TypeHeaderCode
#include "parser_generator.h"
using namespace sscientists::parsejoy;

%End

public:
    ParserGenerator();
    void setPlugin(ParserPlugin *p);
};
