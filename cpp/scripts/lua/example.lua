yaml = parsejoy.loadYAML("../examples/python/grammar.yml")
print("Successfully loaded YAML")
print(yaml)
grammar = parsejoy.parseYAMLGrammar(yaml)
print("Successfully parsed grammar...")
map_rule = parsejoy.toMapRule(grammar)
tokenizer = map_rule.rules:get('tokenizer')
print("Found tokenizer!")
sp = parsejoy.StringParserGenerator(tokenizer, env)
print("Compiling...")
sp.debug = false
parser, err = sp:compile()
print(err)
print("Done!")
input = [[import foo
import bar

def foo(nar):
    pass
    x = a+x

class Foobar(object):
    pass

    def __init__(self, foo : test = "foo"):
        pass
]]

start = os.clock()
for i=1,10 do
    ss = parsejoy.StringState(env, input)
    result = parsejoy.runParser(parser, ss)
end
stop = os.clock()
print(stop-start)
print(ss:Position(),ss:Size())