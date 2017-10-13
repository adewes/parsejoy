#include <string>

using namespace std;

namespace sscientists {
namespace parsejoy {

class Node {
public:
  Node();
  bool __contains__(std::string str);
  int __getitem__(int x) const;
  void __setitem__(int x, int y);
  int operator[](int x);
  ~Node();
  int  length;
};

}
}
