
#include "lua.hpp"
#include "lualib.h"
#include "lauxlib.h"
#include "lua_environment.h"
#include <iostream>
#include <fstream>
#include <streambuf>
#include <sys/types.h>
#include <pwd.h>
#include <unistd.h>
#include "lua_runtime.h"

extern "C"
{
    #include "c/prompt.h"
    #include <string.h>
    #include <stdio.h>

    int luaopen_parsejoy(lua_State* L); // declare the wrapped module

    typedef struct {
        size_t size;
        byte_t* p;
    } lua_data;

    int writer(lua_State* L, const void* p, size_t sz, void* ud){
        lua_data* ld = (lua_data *)ud;
        ld->size += sz;
        ld->p = (byte_t *)realloc(ld->p, ld->size);
        memcpy(ld->p+ld->size-sz, p, sz);
        return 0;
    }

}


namespace sscientists {
namespace parsejoy {

void* LuaEnvironment::extract(const std::string name, const std::string typeName, bool transferMemory){
    auto typeInfo = SWIG_TypeQuery(l_,(typeName+" *").c_str());
    if (typeInfo == NULL){
        throw std::runtime_error("Unknown type name: "+typeName);
    }
    lua_getglobal(l_, name.c_str());
    if (lua_isuserdata(l_, -1)){
        swig_lua_userdata* data = (swig_lua_userdata*)lua_touserdata(l_,-1);
        lua_pop(l_, 1);
        if (data->type != typeInfo)
            return nullptr;//wrong object type!
        //we transfer the responsibility of the memory to the C++ code
        if (transferMemory)
            data->own = 0;
        return data->ptr;
    }
    lua_pop(l_,1);
    return nullptr;
}

void LuaEnvironment::place(const std::string name,void* p, const std::string typeName, bool manageMemory){
    /*
    Creates a global variable containing a reference to the given LuaEnvironment
    variable (passed as a void* pointer). Assumes that the memory management
    of the
    */
    //creates a global variable containing a reference to the given LuaEnvironment

    auto typeInfo = SWIG_TypeQuery(l_,(typeName+" *").c_str());
    if (typeInfo == NULL){
        throw std::runtime_error("Unknown type name: "+typeName);
    }
    SWIG_NewPointerObj(l_,p,typeInfo,manageMemory ? 1: 0);
    lua_setglobal(l_, name.c_str());
}

LuaEnvironment::LuaEnvironment(){
    l_ = luaL_newstate();
    luaopen_base(l_);
    luaL_openlibs(l_);
    luaopen_parsejoy(l_);
    place("env", this, "sscientists::parsejoy::LuaEnvironment", false);
}

LuaEnvironment::~LuaEnvironment(){
    lua_close(l_);
}

int LuaEnvironment::runEvalLoop(){
    const char *homedir;
    if ((homedir = getenv("HOME")) == NULL) {
        homedir = getpwuid(getuid())->pw_dir;
    }
    std::string historyFile = homedir;
    historyFile+="/.parsejoy-history";
    luap_sethistory(l_, historyFile.c_str());
    luap_enter(l_);
    return 0;
}

LuaByteCode LuaEnvironment::compileCode(const std::string code){
    auto result = luaL_loadstring(l_, code.c_str());
    switch (result){
        case LUA_ERRMEM:
            throw std::runtime_error("Cannot allocate memory for code!");
        case LUA_ERRSYNTAX:
            size_t n;
            auto error = lua_tolstring (l_, -1, &n);
            std::string message = "Syntax error: \n";
            message += error;
            throw std::runtime_error(error);
    }
    lua_data ld;
    ld.size = 0;
    ld.p = (byte_t *)malloc(1);
    lua_dump(l_, writer, (byte_t *)&ld);

    LuaByteCode byteCode;
    byteCode.code = shared_ptr<byte_t>(ld.p);
    byteCode.size = ld.size;
    return byteCode;
}

int LuaEnvironment::copyUserData(int r){
    if (r == LUA_NOREF)
        return r;
    lua_getglobal(l_, "copy");
    lua_rawgeti(l_, LUA_REGISTRYINDEX, r);
    if (lua_pcall(l_, 1, 1, 0) != 0){
        size_t n;
        const char* message = lua_tolstring(l_, -1, &n);
        throw std::runtime_error("Unable to copy Lua state: "+(message != nullptr ? std::string(message) : "(no exception information given)"));
    }
    return (int) luaL_ref(l_, LUA_REGISTRYINDEX);
}

int LuaEnvironment::runScript(const std::string filename){
    try{
        std::ifstream t(filename);
        std::string content((std::istreambuf_iterator<char>(t)),
                         std::istreambuf_iterator<char>());
        return runCode(content);
    }catch(const std::exception& e){
        throw;
    }
}

std::string LuaEnvironment::getError(){
    size_t n;
    auto ptr = lua_tolstring(l_, -1, &n);
    if (ptr == nullptr)
        return "";
    return std::string(ptr);
}

int LuaEnvironment::runCode(const std::string code){
    return luaL_dostring(l_, code.c_str());
}

int LuaEnvironment::runByteCode(LuaByteCode byteCode){
    luaL_loadbuffer(l_, (const char*)byteCode.code.get(), byteCode.size,"bytecode");
    lua_pcall(l_, 0, LUA_MULTRET, 0);
}

}
}
