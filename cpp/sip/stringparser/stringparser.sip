%Module(name=parsejoy.stringparser,version=1,call_super_init=False)

%Import ../parser_generator/parser_generator.sip

class StringParserPlugin : public ParserPlugin {

%TypeHeaderCode
#include "stringparser.h"

using namespace sscientists::parsejoy;

%End

public:
    StringParserPlugin();
};
