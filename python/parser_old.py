import yaml
import sys
import os
import time
import re
import copy
import pprint

"""
For each possible rule path
"""

class ParserError(ValueError):
    pass

class ParserError(ValueError):
    pass


class Context(object):

    def __init__(self,level,name,parent=None):
        self.level = level
        self.name = name
        self.parent = parent
        if self.parent:
            self.url = self.parent.url+'.'+self.name
        else:
            self.url = self.name

    def debug(self,msg):
        return
        print "{}{}: {}".format(" "*self.level,self.name,msg)

class State(object):

    def __init__(self,s, pos=0):
        self.s = s
        self.parent = None
        self.store = {}
        self.result = None
        self.current_node = []
        self.root = self.current_node
        self.pos = pos

    @property
    def line(self):
        return len(self.s[:self.pos].split("\n"))

    @property
    def col(self):
        return len(self.s[:self.pos].split("\n")[-1])+1

    def copy(self,):
        state = State(self.s, self.pos)
        state.parent = self
        state.store = copy.deepcopy(self.store)
        state.current_node = self.current_node
        state.root = self.root
        return state

    @property
    def value(self):
        return self.s[self.pos:]

    def advance(self, n):
        old_pos = self.pos
        self.pos += n
        return old_pos

    def go_to(self, pos):
        old_pos = self.pos
        self.pos = pos
        return old_pos

from collections import defaultdict
encountered_contexts = defaultdict(dict)

class Iterator(object):

    def __init__(self,generator, parent=None):
        self.generator = generator
        self.parent = parent
        self.list = []
        self.pos = 0

    def __iter__(self):
        return self

    def get(self,pos):
        if self.parent:
            return self.parent.get(pos)
        while pos >= len(self.list):
            value = next(self.generator)
            self.list.append(value)
        return self.list[pos]

    def next(self):
        self.pos+=1
        return self.get(self.pos-1)

    def copy(self):
        if self.parent:
            return Iterator(None,parent=self.parent)
        return Iterator(None,parent=self)


def parser(name, url):

    """
    Simplifying parser rules

    * or
    """

    def dec(f):
        return f
        def decorated_function(state, context, *args, **kwargs):
            if False and url is not None and url in encountered_contexts[state.pos]:
#                print url,state.pos
                return encountered_contexts[state.pos][url].copy()
            print("{}{} {}:{}".format(" "*context.level,context.name,state.line,state.col))
            new_context = Context(context.level+1,name,context)
            result = f(state, new_context, *args, **kwargs)
            if url is not None:
                encountered_contexts[state.pos][url] = Iterator(result)
                return encountered_contexts[state.pos][url]
            return result
        return decorated_function

    return dec

class ParserGenerator(object):

    """
    Generating an abstract syntax tree is done implictly by each rule

    """

    def __init__(self, grammar):
        self.grammar = grammar
        self.parsers = {}

    def compile_regex(self, regex, url):
        
        compiled_regex = re.compile('^{}'.format(regex))

        @parser('regex', url)
        def regex_parser(state, context):
            context.debug(regex)
            match = compiled_regex.match(state.value)
            if match:
                s = match.group(0)
                context.debug("match!")
                new_state = state.copy()
                new_state.result = s
                new_state.advance(len(s))
                yield new_state
            else:
                raise ParserError("Regex not matched: {}".format(regex))

        return regex_parser

    def compile_ref(self, key, url):

        def ref_parser(state):
            new_state = state.copy()
            new_state.result = state.store.get(key)
            yield new_state

        return ref_parser

    def compile_ast_list(self, props, url):

        name = props.get('name')
        rule_parser = self._compile_rule(props['value'], url+'.ast-list')

        @parser('ast-list', url)
        def ast_list_parser(state, context):

            l = []

            current_node = state.current_node
            state.current_node = l
            try:
                for new_state in rule_parser(state, context):

                    if isinstance(current_node,dict) and name:
                        new_current_node = new_state.current_node
                        new_state.current_node = current_node.copy()
                        new_state.current_node[name] = new_current_node

                    yield new_state
            finally:
                state.current_node = current_node


        return ast_list_parser


    def compile_ast_prop(self, props, url):

        name = props.get('name')
        value_parser = self._compile_rule(props['value'], url+'.ast-prop')

        @parser('ast-prop', url)
        def ast_prop_parser(state, context):

            for new_state in value_parser(state, context):
                current_node = new_state.current_node
                if isinstance(current_node,dict):
                    current_node[name] = new_state.result
                yield new_state

        return ast_prop_parser

    def compile_ast_node(self, props, url):

        """
        Create a new AST node.

        * If the current node is a list, appends the new node to it
        * If the current node is a dict, puts the new node in the key given by name (if provided)
        * If none of these things match, does nothing
        """

        rule_parser = self._compile_rule(props['value'], url+'.ast-node')
        name = props.get('name')

        @parser('ast-node', url)
        def ast_node_parser(state, context):

            d = {}
            d.update(props.get('props',{}))

            current_node = state.current_node
            state.current_node = d
            try:
                for new_state in rule_parser(state, context):
                    new_current_node = new_state.current_node
                    if isinstance(current_node,list):
                        new_state.current_node = current_node[:]
                        new_state.current_node.append(new_current_node)
                    elif isinstance(current_node,dict):
                        new_state.current_node = current_node.copy()
                        if name:
                            new_state.current_node[name] = new_current_node
                        else:
                            new_state.current_node.update(new_current_node)

                    yield new_state
            finally:
                state.current_node = current_node

        return ast_node_parser

    def compile_repeat(self, rule, url):

        rule_parser = self._compile_rule(rule, url+'.repeat')

        @parser('repeat', url)
        def repeat_parser(state, context):
            cnt=0
            current_state = state
            states_to_repeat=[state]
            states_to_yield = []
            productions = []
            while states_to_repeat or states_to_yield or productions:
                if states_to_repeat:
                    current_state=states_to_repeat.pop()
                    states_to_yield.append(current_state)
                    try:
                        production=rule_parser(current_state, context)
                        new_state = next(production)
                        #if the production does not advance the state, we reject it...
                        if new_state.pos == current_state.pos:
                            continue
                        productions.append(production)
                        states_to_repeat.append(new_state)
                    except (ParserError, StopIteration) as e :
                        continue
                elif states_to_yield:
                    state_to_yield = states_to_yield.pop()
                    cnt +=1
                    if state_to_yield != state:
                        yield state_to_yield
                elif productions:
                    production = productions[-1]
                    try:
                        new_state = next(production)
                        states_to_yield.append(new_state)
                    except (ParserError,StopIteration):
                        productions.pop()
            if cnt==0:
                raise ParserError("Not matched!")

        return repeat_parser

    def compile_optional(self, rule, url):

        rule_parser = self._compile_rule(rule, url+'.optional')

        @parser('optional', url)
        def optional_parser(state, context):
            try:
                for new_state in rule_parser(state, context):
                    yield new_state
            except ParserError as me:
                pass
            yield state

        return optional_parser

    def compile_store(self, args, url):

        name = args['name']
        value = args['value']

        value_parser = self._compile_rule(value, url+'.store')

        @parser('store', url)
        def store_parser(state, context):
            for ns in value_parser(state, context):
                new_state = state.copy()
                new_state.result = ns.result
                yield new_state

        return store_parser

    def compile_literal(self, value, url):

        if isinstance(value, dict):
            value = self._compile_rule(value, url+'.literal')

        @parser('literal', url)
        def literal_parser(state, context):
            context.debug(value)
            if callable(value):
                v = value(state, context)
            else:
                v = value
            found_value = state.value[:len(v)]
            if found_value != v:
                raise ParserError("Expected {}, but found '{}'".format(value, found_value))
            context.debug(v)
            new_state = state.copy()
            new_state.advance(len(v))
            new_state.result = v
            yield new_state

        return literal_parser

    def compile_python_code(self, code, url):

        gv = globals().copy()
        gv['url'] = url
        exec(code,gv,gv)
        return gv['parser']

    def compile_or(self, alternatives, url):

        alternative_parsers = []
        for i,alternative in enumerate(alternatives):
            alternative_parsers.append((alternative,self._compile_rule(alternative, url+'.or.{}'.format(i))))

        @parser('or', url)
        def or_parser(state, context):
            """
            Pass in context object that contains information about the following things:

            * Which rule has called this one?
            *
            """
            found = False
            alternative_productions = []
            for params,alternative_parser in alternative_parsers:
                try:
                    alternative_productions.append(alternative_parser(state, context))
                except ParserError as me:
                    continue
            i = 0
            while alternative_productions:
                production = alternative_productions[i%len(alternative_productions)]
                try:
                    new_state = next(production)
                    found = True
                    yield new_state
                    i+=1
                except (ParserError,StopIteration):
                    alternative_productions.remove(production)
            if not found:
                raise ParserError("No alternative matched!")

        return or_parser

    def compile_sequence(self, rules, url):

        """
        Increase the level by one for each element in the sequence
        """

        parsers = []

        for i,rule in enumerate(rules):
            ps = self._compile_rule(rule, url+'.seq.{}'.format(i))
            if ps is None:
                raise AttributeError
            parsers.append(ps)

        @parser('sequence', url)
        def sequence_parser(state, context):
            """
            * Execute the first parser on the state
            * For each returned state, execute the second parser
            * For each returned state, execute the third parser...
            """

            def parse_sequence(state, parsers):
                parser = parsers.pop(0)
                for new_state in parser(state, context):
                    if parsers:
                        try:
                            for new_new_state in parse_sequence(new_state, parsers[:]):
                                yield new_new_state
                        except ParserError:
                            continue
                    else:
                        yield new_state


            for new_state in parse_sequence(state, parsers[:]):
                yield new_state

        return sequence_parser

    def compile(self, debug=True):
        self.parsers = {}
        return self._compile_rule('start', '')

    def _compile_rule(self, name_or_rule, url):
        """
        Takes a YAML grammar as input and returns a Python parser function that can be
        called with a Stream instance and a state as arguments.
        """

        name = None

        if isinstance(name_or_rule,(str,unicode)):
            name = name_or_rule
            if name in self.parsers:
                return self.parsers[name]
            rule = self.grammar[name]
        else:
            rule = name_or_rule

        if name:
            new_url = url+'.'+name
        else:
            new_url = url

        def parse_subrule(rule, name=None):
            rule_name = rule.keys()[0]
            args = rule.values()[0]
            if rule_name == '$python':
                result = self.compile_python_code(args, url+'.{}'.format(name))
                if name:
                    self.parsers[name] = result
                return result
            try:
                func = getattr(self,'compile_{}'.format(rule_name.replace('-','_')))
            except AttributeError:
                raise ParserError("Unknown rule: {}".format(rule_name))

            subparser = func(args, new_url)

            @parser(rule_name, None)
            def subrule_parser(state, context):
                for result in subparser(state, context):
                    yield result

            if name:
                @parser(name, None)
                def name_parser(state, context):
                    for result in subrule_parser(state, context):
                        yield result
                self.parsers[name] = name_parser
                return name_parser
            return subrule_parser

        #this allows definition of recursive parsing rules via a simple function call
        if name:
            #this will lead to infinite recursion if the parser is not replaced!

            @parser(name, url)
            def subrule_parser(state, context):
                for result in self.parsers[name](state, context):
                    yield result

            self.parsers[name] = subrule_parser

        if isinstance(rule,(list,tuple)):
            sequence_parser = self.compile_sequence(rule, new_url)
            if name:
                @parser(name, None)
                def subrule_parser(state, context):
                    for result in sequence_parser(state, context):
                        yield result
                self.parsers[name] = subrule_parser
                return subrule_parser
            return sequence_parser
        elif isinstance(rule,dict) and len(rule) == 1:
            return parse_subrule(rule, name=name)
        elif isinstance(rule,(str,unicode)):

            new_new_url = new_url+'.'+rule
            ps = self._compile_rule(rule, new_new_url)

            @parser(name, None)
            def subrule_parser(state, context):
                for result in ps(state, context):
                    yield result

            self.parsers[name] = subrule_parser

            return subrule_parser

        raise ParserError("Unknown rule: {}".format(name or name_or_rule or '(no name given)'))


if __name__ == '__main__':
    import sys
    sys.setrecursionlimit(100000)
    if len(sys.argv) < 3:
        sys.stderr.write("Usage: {} [grammar filename] [code filename]\n".format(os.path.basename(__file__)))
        exit(-1)
    grammar_filename = sys.argv[1]
    code_filename = sys.argv[2]

    with open(grammar_filename,'r') as grammar_file:
        grammar = yaml.load(grammar_file.read())

    with open(code_filename,'r') as code_file:
        code = code_file.read()

    parser_generator = ParserGenerator(grammar)

    parser = parser_generator.compile()
    state = State(code)

    start = time.time()

    results = parser(state, Context(0,'root',None))

    for result in results:

        print result.line,result.col
        if result.value.strip():
            print "Parsing failed in line {}, column {}:\n\n{}...".format(result.line,result.col,result.value[:20])
        else:
            print "Parsing succeeded!"
            pprint.pprint(result.current_node)

