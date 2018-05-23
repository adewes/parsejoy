--we load the Python YAML grammar
yaml = parsejoy.loadYAML("../examples/calculator/grammar.yml")
print("Successfully loaded YAML")
print(yaml)
print("Parsing grammar...")
--we parse the grammar
grammar = parsejoy.parseYAMLGrammar(yaml)
print("Successfully parsed grammar...")

--we convert it to a map rule
map_rule = parsejoy.toMapRule(grammar)
print("Generating string parser generator...")
--we generate a string parser generator with the tokenizer
sp = parsejoy.StringParserGenerator(grammar, env)
print("Compiling...")
sp.debug = false

--we compile the parser generator
parser, err = sp:compile()
print("Error:",err)
print("Done compiling grammar...")

--a simple example Python script
input = [[1+1+4+4*(5+5*(5+5)*(6+6*5))+1+1+1+1+1+1+1111111+1+1+1+111123+1212-1]]

print("Parsing the input 10 times...")

--we benchmark the parsing code
start = os.clock()
for i=1, 10 do
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
