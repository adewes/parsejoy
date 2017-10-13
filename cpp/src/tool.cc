#include <iostream>
#include <string>

#include "stringparser.h"
#include "grammar.h"
#include "lua_environment.h"
#include "set.h"

using namespace sscientists::parsejoy;
using namespace std;

class InputParser{
    public:
        InputParser (int &argc, char **argv){
            for (int i=1; i < argc; ++i)
                this->tokens.push_back(std::string(argv[i]));
        }
        /// @author iain
        const std::string getCmdOption(const std::string &option) const{
            std::vector<std::string>::const_iterator itr;
            itr =  std::find(this->tokens.begin(), this->tokens.end(), option);
            if (itr != this->tokens.end() && ++itr != this->tokens.end()){
                return *itr;
            }
            return "";
        }
        /// @author iain
        bool cmdOptionExists(const std::string &option) const{
            return std::find(this->tokens.begin(), this->tokens.end(), option)
                   != this->tokens.end();
        }
    private:
        std::vector <std::string> tokens;
};

int main(int argc, char* argv[]) {
    InputParser input(argc, argv);
    auto luaEnvironment = LuaEnvironment();
    luaEnvironment.runScript("scripts/lua/setup.lua");
    const std::string &filename = input.getCmdOption("-f");
        if (!filename.empty()){
            if (luaEnvironment.runScript(filename) != 0){
                cout << luaEnvironment.getError() << "\n";
            }
        }
    luaEnvironment.runEvalLoop();
}
