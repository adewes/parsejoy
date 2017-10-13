--we load the Python YAML grammar
yaml = parsejoy.loadYAML("../examples/python/grammar.yml")
print("Successfully loaded YAML")
print(yaml)

--we parse the grammar
grammar = parsejoy.parseYAMLGrammar(yaml)
print("Successfully parsed grammar...")

--we convert it to a map rule
map_rule = parsejoy.toMapRule(grammar)

--we extract the tokenizer
tokenizer = map_rule.rules:get('tokenizer')
print("Found tokenizer!")

--we generate a string parser generator with the tokenizer
sp = parsejoy.StringParserGenerator(tokenizer, env)
print("Compiling...")
sp.debug = false

--we compile the parser generator
parser, err = sp:compile()
print("Error:",err)
print("Done compiling grammar...")

--a simple example Python script
input = [[import foo
import bar

def foo(nar):
    pass
    x = a+x
--- : 
class Foobar(object):
    pass

    def __init__(self, foo : test = "foo"):
        pass
]]

print("Parsing the input 10 times...")

--we benchmark the parsing code
start = os.clock()
for i=1, 100 do
    ss = parsejoy.StringState(env, input)
    result = parsejoy.runParser(parser, ss)
end
stop = os.clock()

print("Elapsed time:",stop-start)

--we check if the parsing was successful
if ss:Position() == ss:Size() then
    print("Successfully parsed the input!")
end

print(ss:Position())