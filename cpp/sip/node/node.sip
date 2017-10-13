%Module(name=parsejoy.node,version=1,call_super_init=False)

%Include ../helpers.sip

class Node {

%TypeHeaderCode
#include "node.h"
using namespace sscientists::parsejoy;
%End

public:
    Node();
    int __getitem__(int i);
    void __setitem__(int i, int j);
    %MethodCode
            sipCpp->__setitem__(a0, a1);
    %End
    int __contains__(std::string str);
    %MethodCode
            sipRes = sipCpp->__contains__(*a0);
    %End
    int length;
};
