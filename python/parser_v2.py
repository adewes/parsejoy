import yaml
import sys
import os
import time
import re
import copy
import pprint
import hashlib

class ParserError(ValueError):
    pass

class ExpectedParserError(ParserError):
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

class TokenStream(object):

    """
    A TokenStream object delivers a stream of tokens to the parser.
    """

    def __init__(self,tokens):
        self.tokens = tokens
        self.current_token = None
        self.consumed_tokens = []
        self.initialized = False

    def build_linked_list(self):
        """
        Build a linked list of tokens:

        * The next token is either the next token in the current list, or the next
          token in the parent token's list.
        """
        print "Building linked list..."
        self.initialized = True

        def add_token_list(token_list,parent,next_parent,previous_parent,token_id=0):
            previous_token = previous_parent
            wrapped_tokens = []
            for i,token in enumerate(token_list):
                wrapped_tokens.append({
                    'value' : token,
                    'parent' : parent,
                    'token_id' : token_id
                    })
                token_id+=1
            for token_1,token_2 in zip(wrapped_tokens[:-1],wrapped_tokens[1:]):
                token_1['next'] = token_2
                token_2['previous'] = token_1
            wrapped_tokens[-1]['next'] = next_parent
            wrapped_tokens[0]['previous'] = previous_parent
            for i,token in enumerate(wrapped_tokens):
                if token['value'].get('c'):
                    next_token = wrapped_tokens[i+1] if i < len(wrapped_tokens)-1 else next_parent
                    token['children'],token_id = add_token_list(token['value']['c'],token,next_token,previous_token,token_id=token_id)
                previous_token = token
            return wrapped_tokens[0],token_id

        self.current_token,token_id = add_token_list(self.tokens, None, None, None,0)


    def copy(self):
        stream = self.__class__(self.tokens)
        stream.current_token = self.current_token
        stream.initialized = self.initialized
        stream.consumed_tokens = self.consumed_tokens[:]
        return stream

    def advance(self, from_token = None, prefix=None, skip_ignored=True):
        if self.current_token is None:
            if not self.initialized:
                self.build_linked_list()
                self.initialized = True
            else:
                raise StopIteration
        if from_token:
            self.current_token = from_token
        if prefix:
            self.consumed_tokens.extend(prefix)
        self.consumed_tokens.append(self.current_token)
        self.current_token = self.current_token.get('next')

    def get(self, token_type=None, leafs_only=False, include_ignored=False):

        if self.current_token is None:
            if not self.initialized:
                self.build_linked_list()
                self.initialized = True
            else:
                return None,[]

        def get_next_relevant_token(token):
            prefix = []
            if include_ignored:
                return token,prefix
            while token['value'].get('ignore') and token['next']:
                prefix.append(token)
                token = token['next']
            return token,prefix

        current_token,prefix = get_next_relevant_token(self.current_token)
        if token_type is not None:
            while current_token['value']['type'] != token_type:
                if current_token.get('children'):
                    current_token,child_prefix = get_next_relevant_token(current_token['children'])
                    prefix.extend(child_prefix)
                else:
                    raise StopIteration
            return current_token,prefix
        elif leafs_only:
            while current_token.get('children'):
                current_token,child_prefix = get_next_relevant_token(current_token['children'])
                prefix.extend(child_prefix)
        return current_token,prefix

    def current(self):
        pass

class State(object):

    def __init__(self):
        self.parent = None
        self.result = None
        self.store = {}
        self.tokens = []

class StringState(State):

    def __init__(self, s, pos=0):
        super(StringState, self).__init__()
        self.s = s
        self.pos = pos

    @property
    def line(self):
        return len(self.s[:self.pos].split("\n"))

    @property
    def col(self):
        return len(self.s[:self.pos].split("\n")[-1])+1

    def copy(self,):
        state = self.__class__(self.s, self.pos)
        state.parent = self
        state.store = copy.deepcopy(self.store)
        state.tokens = self.tokens[:]
        return state

    def print_parse_tree(self, tokens=None, level=0):
        if tokens is None:
            tokens = self.tokens
        for token in tokens:
            start = token['from']['p']
            stop = token['to']['p']
            print "{}{} ({}-{})".format(" "*level,token['type'],start,stop)
            if token.get('c'):
                self.print_parse_tree(token['c'],level=level+1)

    @property
    def value(self):
        return self.s[self.pos:]

    def advance(self, n):
        old_pos = self.pos
        self.pos += n
        return old_pos

    def create_token(self, name, s, push=True, **opts):
        token = {
          'type' : name,
          's' : s,
          'from' : {'p' :self.pos,'l' : self.line,'c' : self.col},
        }
        if name.startswith('__'):
            token['ignore'] = True
        token.update(opts)
        old_pos = self.advance(len(s))
        token['to'] = {'p' : self.pos,'l' : self.line,'c' : self.col}
        self.go_to(old_pos)
        if push:
            self.tokens.append(token)
        return token

    def go_to(self, pos):
        old_pos = self.pos
        self.pos = pos
        return old_pos

class TokenState(State):

    def __init__(self, tokens):
        super(TokenState,self).__init__()
        self.s = tokens

    def print_parse_tree(self, tokens=None, level=0):
        if tokens is None:
            tokens = self.tokens
        for token in tokens:
            if token['tokens']:
                start = token['tokens'][0]['value']['from']['p']
                stop = token['tokens'][-1]['value']['to']['p']
            else:
                start = ''
                stop = ''
            print "{}{} ({}-{})".format(" "*level,token['type'],start,stop)
            if token.get('c'):
                self.print_parse_tree(token['c'],level=level+1)

    def copy(self,):
        state = TokenState(self.s.copy())
        state.parent = self
        state.tokens = self.tokens[:]
        return state

    def create_token(self, name, tokens, push=True, **opts):
        token = {
          'type' : name,
          'tokens' : tokens
        }
        token.update(opts)
        if push:
            self.tokens.append(token)
        return token

    def current_token(self):
        return self.s.current_token

    def value(self, token_type=None):
        return self.s.get(token_type=token_type, leafs_only=True if token_type is None else False)

    def advance(self, *args, **kwargs):
        return self.s.advance(*args, **kwargs)

class BaseParserGenerator(object):

    def parser(self, f, *args, **kwargs):
        raise NotImplementedError

    def get_fingerprint(self,rule):
        hasher = hashlib.sha1()
        if isinstance(rule,(str,unicode)):
            hasher.update(rule)
        elif isinstance(rule,(list,tuple)):
            for elem in rule:
                hasher.update(self.get_fingerprint(elem))
        elif isinstance(rule,dict):
            for key in sorted(rule.keys()):
                hasher.update(key)
                hasher.update(self.get_fingerprint(rule[key]))
        return hasher.hexdigest()

    def __init__(self, grammar):
        self.grammar = grammar
        self.parsers = {}
        self.prefixes = {}
        for key in self.grammar:
            self.prefixes[key] = self.get_prefixes(key)

    def compile_repeat(self, rule):

        rule_parser = self._compile_rule(rule['repeat'])

        @self.parser('repeat',rule=rule, emit=False)
        def repeat_parser(state, context):
            cnt=0
            current_state = state
            while True:
                try:
                    current_state = rule_parser(current_state, context)
                    cnt+=1
                except ParserError:
                    break
            if cnt == 0:
                raise ParserError
            return current_state

        return repeat_parser

    def compile_optional(self, rule):

        rule_parser = self._compile_rule(rule['optional'])

        @self.parser('optional',rule=rule, emit=False)
        def optional_parser(state, context):
            try:
                return rule_parser(state, context)
            except ParserError as me:
                return state

        return optional_parser

    def compile_or(self, rule):

        alternative_parsers = []
        alternatives = rule['or']
        for i,alternative in enumerate(alternatives):
            alternative_parsers.append((alternative,self._compile_rule(alternative)))

        @self.parser('or',rule=rule, emit=False)
        def or_parser(state, context):
            """
            Pass in context object that contains information about the following things:

            * Which rule has called this one?
            *
            """
            found = False
            for alternative,alternative_parser in alternative_parsers:
                try:
                    return alternative_parser(state, context)
                except ParserError as pe:
                    continue
            else:
                raise ParserError("No alternative matched!")

        return or_parser

    def compile_not(self, rule):

        rule_parser = self._compile_rule(rule['not'])

        @self.parser('not',rule=rule, emit=False)
        def not_parser(state, context):
            try:
                rule_parser(state, context)
            except ParserError:
                return state
            raise ParserError("not condition did match!")

        return not_parser

    def compile_and(self, rule):

        rule_parser = self._compile_rule(rule['and'])

        @self.parser('and',rule=rule, emit=False)
        def and_parser(state, context):
            try:
                rule_parser(state, context)
                return state
            except ParserError:
                raise ParserError("and condition did not match!")

        return and_parser

    def compile_sequence(self, rules):

        """
        Increase the level by one for each element in the sequence
        """

        parsers = []

        for i,rule in enumerate(rules):
            ps = self._compile_rule(rule)
            if ps is None:
                raise AttributeError
            parsers.append((rule,ps))

        @self.parser('sequence', rule=rules,emit=False)
        def sequence_parser(state, context):
            """
            * Execute the first parser on the state
            * For each returned state, execute the second parser
            * For each returned state, execute the third parser...
            """

            current_state = state.copy()
            for rule,parser in parsers:
                current_state = parser(current_state, context)
            return current_state

        return sequence_parser

    def compile(self, debug=True):
        self.parsers = {}
        return self._compile_rule('start')

    def resolve_rule(self,name):
        func = getattr(self,'compile_{}'.format(name.replace('-','_')))
        return func

    def rule_prefix(self,rule_name):
        return set([])

    def can_proceed(self, rule, state):
        return True

    def get_prefixes(self,rule,visited_rules = None):

        if visited_rules is None:
            visited_rules = set()
        prefixes = set()

        get_prefixes = lambda rule: self.get_prefixes(rule,visited_rules=visited_rules)
        if isinstance(rule, (str,unicode)):
            if rule in visited_rules:
                return set([])
            visited_rules|=set([rule])
            if rule in self.grammar:
                return get_prefixes(self.grammar[rule])
            else:
                return self.rule_prefix(rule)
        elif isinstance(rule,(tuple,list)):
            for i,subrule in enumerate(rule):
                subrule_prefixes = get_prefixes(subrule)
                if (not subrule_prefixes and not prefixes) or (None in subrule_prefixes):
                    prefixes |= subrule_prefixes - set([None])
                else:
                    prefixes |= subrule_prefixes
                    break
        elif isinstance(rule,dict) and len(rule) == 1:
            key,value = rule.items()[0]
            if key == 'or':
                for subrule in value:
                    prefixes |= get_prefixes(subrule)
            elif key in ('and','repeat'):
                prefixes |= get_prefixes(value)
            elif key == 'not':
                #not sure if this is the right way to handle not
                return set([None])
#                prefixes |= {('not',prefix) for prefix in get_prefixes(value)}
            elif key == 'optional':
                prefixes |= set([None])
                prefixes |= get_prefixes(value)
            else:
                return self.rule_prefix(rule)
        else:
            raise ValueError("Invalid rule!")


    def _compile_rule(self, name_or_rule):

        name = None

        if isinstance(name_or_rule,(str,unicode)):
            name = name_or_rule
            if name in self.parsers:
                return self.parsers[name]
            try:
                rule = self.grammar[name]
            except KeyError:
                try:
                    func = self.resolve_rule(name)
                    return func()
                except AttributeError:
                    raise ParserError("Unknown rule: {}".format(name))
        else:
            rule = name_or_rule

        def parse_subrule(rule, name=None):
            rule_name = rule.keys()[0]
            try:
                func = self.resolve_rule(rule_name)
            except AttributeError:
                raise ParserError("Unknown rule: {}".format(rule_name))

            subparser = func(rule)

            if name:
                @self.parser(name,rule=name,emit=True)
                def name_parser(state, context):
                    return subparser(state, context)
                self.parsers[name] = name_parser
                return name_parser
            return subparser

        #this allows definition of recursive parsing rules via a simple function call
        if name:
            #this will lead to infinite recursion if the parser is not replaced!

            def subrule_parser(state, context):
                return self.parsers[name](state, context)

            self.parsers[name] = subrule_parser

        if isinstance(rule,(list,tuple)):
            sequence_parser = self.compile_sequence(rule)
            if name:
                @self.parser(name,rule=name,emit=True)
                def subrule_parser(state, context):
                    return sequence_parser(state, context)
                self.parsers[name] = subrule_parser
                return subrule_parser
            return sequence_parser
        elif isinstance(rule,dict) and len(rule) == 1:
            return parse_subrule(rule, name=name)
        elif isinstance(rule,(str,unicode)):
            ps = self._compile_rule(rule)
            @self.parser(name,rule=rule,emit=True)
            def subrule_parser(state, context):
                return ps(state, context)

            self.parsers[name] = subrule_parser

            return subrule_parser

        raise ParserError("Unknown rule: {}".format(name or name_or_rule or '(no name given)'))


class StringParserGenerator(BaseParserGenerator):

    """
    Generating an abstract syntax tree is done implictly by each rule

    """

    def parser(self,name,rule=None, emit=True):

        def dec(f):

            def decorated_function(state, context, *args, **kwargs):
                new_context = Context(context.level+1,name,context)
                #print("{}{} {}:{}".format(" "*new_context.level,new_context.name,state.line,state.col))
                i = len(state.tokens)

                result = f(state, new_context, *args, **kwargs)
                if emit:
                    #this rule was successful, we can create a node in the parse tree for it
                    children = result.tokens[i:]
                    result.tokens = result.tokens[:i]
                    new_token = state.create_token(name,state.value[:result.pos-state.pos],push=False)
                    result.tokens.append(new_token)
                    if children:
                        new_token['c'] = children
                return result
            return decorated_function

        return dec

    def rule_prefix(self,rule):
        if isinstance(rule,(str,unicode)):
            return set([rule])
        elif isinstance(rule,dict) and len(rule) == 1:
            key,value = rule.items()[0]
            if key == 'literal':
                return set([value])
            elif key == 'regex':
                return set([('regex',value)])
        else:
            raise ValueError

    def compile_eof(self):

        @self.parser('eof',rule='eof', emit=True)
        def eof_parser(state, context):
            if state.value != '':
                raise ParserError("Expected EOF")
            return state

        return eof_parser

    def compile_indent(self):

        @self.parser('indent',rule='indent', emit=False)
        def indent_parser(state, context):
            """
            * Match the indentation
            * Compare the matched indentation to the current one.
            * If it is longer, emit a CURRENT_INDENT INDENT token sequence
            * If it is shorter, emit a DEDENT+ CURRENT_INDENT token sequence
            """

            indent_str = re.match(r'^[ \t]*',state.value).group(0)
            new_state = state.copy()
            indents = new_state.store.get('indent',[''])
            current_indent = indents[-1]
            if indent_str == current_indent:
                #yield a CURRENT_INDENT token
                new_state.create_token('current_indent',indent_str,ignore=True)
                new_state.advance(len(indent_str))
                return new_state
            elif len(indent_str) > len(current_indent) and indent_str[:len(current_indent)] == current_indent:
                #this is an indentation
                new_state.create_token('current_indent',current_indent,ignore=True)
                new_state.advance(len(current_indent))
                new_indent = indent_str[len(current_indent):]
                new_state.create_token('indent',new_indent)
                new_state.advance(len(new_indent))
                indents.append(indent_str)
                new_state.store['indent'] = indents
                return new_state
            else:
                #this should be a dedentation
                possible_indents = indents[:-1]
                cnt = 0
                while True:
                    if not possible_indents:
                        raise ParserError("Dedentation does not match!")
                    cnt+=1
                    possible_indent = possible_indents.pop(-1)
                    if possible_indent == indent_str:
                        for i in range(cnt):
                            new_state.create_token('dedent','')
                        new_state.create_token('current_indent',possible_indent,ignore=True)
                        new_state.store['indent'] = possible_indents+[possible_indent]
                        new_state.advance(len(indent_str))
                        return new_state

        return indent_parser

    def compile_regex(self, rule):

        regex = rule['regex']
        #DOTALL is necessary to match newlines
        compiled_regex = re.compile('^{}'.format(regex),re.DOTALL)

        @self.parser('regex', rule=rule)
        def regex_parser(state, context):
            match = compiled_regex.match(state.value)
            if match:
                s = match.group(0)
                context.debug("match!")
                new_state = state.copy()
                new_state.result = s
                new_state.advance(len(s))
                return new_state
            else:
                raise ParserError("Regex not matched: {}".format(regex))

        return regex_parser

    def compile_literal(self, rule):

        value = rule['literal']
        if isinstance(value, dict):
            value = self._compile_rule(value)

        @self.parser('literal',rule=rule)
        def literal_parser(state, context):
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
            return new_state

        return literal_parser

    def resolve_rule(self,name):
        try:
            return super(StringParserGenerator,self).resolve_rule(name)
        except AttributeError:

            def compiler():
                parser = self.compile_literal(rule={'literal':name})

                @self.parser(name,emit=True)
                def wrapped_parser(state, context):
                    return parser(state, context)

                return wrapped_parser
            return compiler

class TokenParserGenerator(BaseParserGenerator):

    def __init__(self, *args, **kwargs):
        super(TokenParserGenerator, self).__init__(*args, **kwargs)
        from collections import defaultdict
        self.failures = 0
        self.outcomes = defaultdict(dict)

    def rule_prefix(self,rule):
        if rule.endswith('!'):
            rule = rule[:-1]
        return set([rule])

    def can_proceed(self, prefixes, state):
        current_token,prelude = state.value()
        tokens = set([])
        while current_token is not None:
            tokens |= set([current_token['value']['type']])
            if current_token['parent'] is not None:
                current_token = current_token['parent']
            else:
                break
        if (tokens & prefixes) or None in prefixes:
            return True
        return False

    def parser(self,name,rule=None,emit=True):

        def dec(f):

            if rule is not None:
                prefixes = self.get_prefixes(rule)
                fingerprint = self.get_fingerprint(rule)
            else:
                prefixes = None
                fingerprint = None

            def decorated_function(state, context, *args, **kwargs):

                new_context = Context(context.level+1,name,context)
                p = state.s.consumed_tokens[-1]['token_id'] if state.s.consumed_tokens else 0

                def print_info():

                    current_token,prefix = state.value()
                    hierarchy = [current_token]
                    cp = current_token
                    while cp['parent'] != None:
                        hierarchy.insert(0,cp['parent'])
                        cp = cp['parent']

                    print("{}{} {}:{}:{}".format(" "*new_context.level,
                                                 new_context.name,
                                                 ','.join([token['value']['type'] for token in hierarchy]),
                                                 current_token['value']['from']['l'],
                                                 current_token['value']['from']['c']))

                if False:
                    print_info()
                if fingerprint and p in self.outcomes and fingerprint in self.outcomes[p]:
                    result = self.outcomes[p][fingerprint]
                    if not isinstance(result,State):
                        raise result
                else:
                    if prefixes is not None and not self.can_proceed(prefixes, state):
                        self.outcomes[p][fingerprint] = ExpectedParserError
                        raise ExpectedParserError
                    try:
                        result = f(state, new_context, *args, **kwargs)
                        self.outcomes[p][fingerprint] = result
                    except ParserError as pe:
                        if not isinstance(pe,ExpectedParserError):
                            #print "Unexpected failure,",name,rule,pe,prefixes
                            #print_info()
                            self.failures += 1
                        self.outcomes[p][fingerprint] = pe
                        raise

                if emit:
                    i = len(state.tokens)
                    consumed_tokens = result.s.consumed_tokens[len(state.s.consumed_tokens):]
                    children = result.tokens[i:]
                    result.tokens = result.tokens[:i]
                    new_token = state.create_token(name,consumed_tokens,push=False)
                    result.tokens.append(new_token)
                    if children:
                        new_token['c'] = children

                return result
            return decorated_function

        return dec

    def compile_token(self, name):
        if name.endswith('!'):
            name = name[:-1]
            breakpoint = True
        else:
            breakpoint = False

        @self.parser(name, rule=name, emit=True)
        def token_parser(state, context):
            new_state = state.copy()
            try:
                token,prefix = new_state.value(token_type=name)
                new_state.advance(from_token=token,prefix=prefix)
                return new_state
            except ValueError:
                token = new_state.value()
                raise ParserError("Expected a token of type {}, but found {} instead".format(name,','.join([t['type'] for t in []])))
            except StopIteration:
                raise ParserError("Expected a token of type {}, but no more tokens found".format(name))
        return token_parser
    
    def resolve_rule(self,name):
        try:
            return super(TokenParserGenerator,self).resolve_rule(name)
        except AttributeError:
            return lambda : self.compile_token(name)

if __name__ == '__main__':
    import sys
    sys.setrecursionlimit(10000)

    if len(sys.argv) < 3:
        sys.stderr.write("Usage: {} [grammar filename] [code filename]\n".format(os.path.basename(__file__)))
        exit(-1)
    grammar_filename = sys.argv[1]
    code_filename = sys.argv[2]

    with open(grammar_filename,'r') as grammar_file:
        grammar = yaml.load(grammar_file.read())

    with open(code_filename,'r') as code_file:
        code = code_file.read()

    if 'tokenizer' in grammar:
        #we tokenize the input, then feed it to the token-based parser
        tokenizer_generator = StringParserGenerator(grammar['tokenizer'])
        tokenizer = tokenizer_generator.compile()
        del grammar['tokenizer']
        parser_generator = TokenParserGenerator(grammar)
        parser = parser_generator.compile()
    else:
        #we directly feed the input to the string-based parser.
        tokenizer = None
        parser_generator = StringParserGenerator(grammar)
        parser = parser_generator.compile()


    start = time.time()

    state = StringState(code)

    if tokenizer:
        tokenizer_result = tokenizer(state,Context(0,'root',None))
        if tokenizer_result.pos != len(code):
            print "Parsing failed!"
            exit(-1)
        token_stream = TokenStream(tokenizer_result.tokens)
        token_stream.build_linked_list()
        token_state = TokenState(token_stream)
        print "Tokenization complete after",time.time()-start
        result = parser(token_state, Context(0,'root',None))
        try:
            if result.value()[0] != None:
                print "Parsing failed (token is not None)!"
                exit(0)
        except (StopIteration):
            pass
    else:
        result = parser(state, Context(0,'root',None))

    stop = time.time()

    while True:
        token,prefix = token_stream.get(leafs_only=True,include_ignored=True)
        if token is None:
            break
        types = []
        ct = token
        while ct is not None:
            types.insert(0,ct['value']['type'])
            ct = ct['parent']
        print "{:4d}-{:4d}:{: <40}  ::: {}".format(token['value']['from']['p'],token['value']['to']['p'],"/".join(types),token['value']['s'])
        try:
            token_stream.advance(from_token=token,skip_ignored=False,prefix=prefix)
        except StopIteration:
            break

    result.print_parse_tree()

    print "Reconstructed source:"

    print "".join([token['value']['s'] for token in result.tokens[0]['tokens']])

    print stop-start

    print parser_generator.failures

