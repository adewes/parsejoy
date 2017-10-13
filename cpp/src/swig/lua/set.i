%include "shared_ptr.i"
%include <std_vector.i>

%{
#include "set.h"
using namespace sscientists::parsejoy;
%}

%include "set.h"

%catches(runtime_error) BitGrammar<string>::AddAs;

%template(StringSet) Set<string>;
%template(StringHashSet) HashSet<string>;
%template(StringBitGrammar) BitGrammar<string>;
%template(StringBitSet) BitSet<string>;
%template(StringVector) std::vector<string>;

%header{

namespace sscientists{
namespace parsejoy{

using namespace std;

shared_ptr<BitGrammar<std::string>> makeStringBitGrammar(){
    return make_shared<BitGrammar<std::string>>();
};
    
}
}

}

shared_ptr<BitGrammar<std::string>> makeStringBitGrammar();

wrap_shared_ptr(StringBitGrammarSharedPtr, BitGrammar<std::string>);
