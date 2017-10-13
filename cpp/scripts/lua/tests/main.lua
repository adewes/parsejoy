--you need to execute this from the "scripts/lua" directory
luaunit = require "luaunit"
require "tests.bitgrammar"
require "tests.bitset"
luaunit.LuaUnit.run()
