#include "node.h"
#include <iostream>

using namespace std;

namespace sscientists {
namespace parsejoy {

Node::Node() {
    cout << "Initializing node\n";
}

Node::~Node() {
    cout << "Destroying node\n";
}

bool Node::__contains__(std::string str) {
    if (str == std::string{"hoo"})
        return true;
    return false;
}

int Node::__getitem__(int x) const {
    cout << "Getting item...\n";
    return this->length;
}

void Node::__setitem__(int x, int y){
    this->length = y;
}

int Node::operator[](int x) {
    return this->length;
}


}
}
