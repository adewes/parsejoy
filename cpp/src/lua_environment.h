#pragma once

#include <string>
#include <memory>
#include "lua.h"
#include "lauxlib.h"

using namespace std;

extern "C" {
    typedef char byte_t;
}

namespace sscientists {
namespace parsejoy {

typedef struct {
    std::shared_ptr<byte_t> code;
    size_t size;
} LuaByteCode;

class LuaEnvironment {
public:
    LuaEnvironment();
    ~LuaEnvironment();
    int copyUserData(int r);
    int runScript(const std::string filename);
    int runCode(const std::string code);
    int runByteCode(LuaByteCode code);
    int runEvalLoop();
    void place(const std::string name, void* p, const std::string typeName, bool manageMemory);
    void* extract(const std::string name, const std::string typeName, bool transferMemory);
    std::string getError();
    LuaByteCode compileCode(const std::string code);
private:
    lua_State *l_;
};


}
}
