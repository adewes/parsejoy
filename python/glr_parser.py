from collections import defaultdict
import time
import re

e_grammar = [
    ['S','E','$'],#0
    ['E','T','+','T'],#1
    ['E','T'],#2
    ['T','E'],#3
    ['T','n'],#4
]

def regex(pattern):

    expr = re.compile(pattern)

    def match(tokens):
        i = 0
        match = re.match(expr, tokens)
        if match:
            return match[0]
        return None
    return match

def literal(value):
    def match(tokens):
        i = 0
        while i < len(tokens) and i < len(value) and tokens[i] == value[i]:
            i += 1
        return tokens[:i]
    return match

gospel_grammar = [
    ['S', 'optional-whitespace', 'statements', '$'],
    ['statements'],
    ['statements', 'statements', 'statement', 'optional_newline'],
    ['optional_newline'],
    ['optional_newline', 'newline'],
    ['newline', '\n'],
    ['statement', 'html', 'whitespace', 'html-statement', 'optional-whitespace'],
    ['html-statement', 'template', 'whitespace', 'template-statement'],
    ['template-statement', 'template-tag'],
    ['template-tag', 'tag-name', '(', 'tag-args', ')'],
    ['tag-name', 'div'],
    ['tag-args',],
    ['optional-whitespace'],
    ['optional-whitespace', 'whitespace'],
    ['whitespace', 'whitespace', 'ws-character'],
    ['whitespace', 'ws-character'],
    ['ws-character', 'comment',],
    ['ws-character', ' '],
    ['ws-character', '\t'],
    ['ws-character', '\n'],
    ['comment', '#', 'anything', 'newline'],
    ['anything', regex('[^\n]*')],
    ['template', literal('template')],
    ['html', literal('html')],
    ['div', literal('div')],
]

gospel_program = [
    'html', 'template', 'div', '(', ')', '\n', 
]

gospel_program = """#abcacbcc 2342342"§$"§$
html # abc
    template # aaaaaaaaaaaaaaaaaaaaaaaaaaaaa
        div() # bbbbbbbbbbb aaaaaaaaaaaaaa ccccc

html template div()
"""

python_grammar = [
    ['S','statements','$'],
    ['statements'], # statements can be empty
    ['statements','statements','statement','optional_newline'],
    ['optional_newline'],
    ['optional_newline','newline'],
    ['statement','funcdef'],
    ['statement','assignment'],
    ['statement','expression'],
    ['statement','pass'],
    ['assignment','name','=','expression'],
    ['funcdef','def','name','parameters',':','suite'],
    ['suite','newline','indent','statements','dedent'],
    ['suite','statement'],
    ['parameters','(','params','keyword_params','optional_comma',')'],
    ['keyword_params'],
    ['optional_comma',],
    ['optional_comma',','],
    ['keyword_params',',','keyword_param','keyword_params'],
    ['keyword_param','name','=','expression'],
    ['expression','name'],
    ['expression','string'],
    ['string','"','characters','"'],
    ['string',"'",'characters',"'"],
    ['expression','number'],
    ['params'],
    ['params','param'],
    ['params','param',',','params'],
    ['param','name'],
]

a_grammar = [
    ['S','A','$'],
    ['A','a','A','a'],
    ['A','b','A','b'],
    ['A','c','A','c'],
    ['A','d','A','d'],
    ['A','e','A','e'],
    ['A','f','A','f'],
    ['A','g','A','g'],
    ['A','h','A','h'],
    ['A','m','A','m'],
    ['A','n','A','n'],
    ['A']
]

term_grammar = [
    ['S','term','$'],
    ['term','factor'],
    ['term','factor','+','term'],
    #['times'],
    #['times','*'],
    ['factor','s','times','factor'],
    ['factor','s'],
    ['s','number'],
    ['s','symbol'],
    ['number', 'digits'],
    ['digits', 'digits', 'digit'],
    ['digits', 'digit'],
    ['digit', '1'],
    ['digit', '2'],
    ['digit', '3'],
]

e_grammar = [
    ['S','e','$'],
    ['e','e','+','e'],
    ['e','b']
]

class Parser(object):

    def __init__(self, grammar, debug=False):
        self.grammar = grammar
        self.debug = debug
        self.generate_states_and_transitions()

    def get_rules_for_non_terminal(self, non_terminal):
        matching_rules = []
        for i,rule in enumerate(self.grammar):
            if rule[0] == non_terminal:
                matching_rules.append(i)
        return matching_rules

    def get_closure(self, rule, pos):
        closure = set()
        rules_to_examine = []
        if pos >= len(self.grammar[rule])-1:
            return closure
        item = self.grammar[rule][pos+1]
        if item in self.non_terminals:
            rules_for_non_terminal = self.get_rules_for_non_terminal(item)
            if self.debug:
                print(item, rules_for_non_terminal)
            closure = set(rules_for_non_terminal)
            rules_to_examine = rules_for_non_terminal
        while rules_to_examine:
            r = rules_to_examine.pop()
            if len(self.grammar[r]) == 1:
                continue
            if self.grammar[r][1] in self.non_terminals:
                new_rules = self.get_rules_for_non_terminal(self.grammar[r][1])
                rules_to_examine.extend([rr for rr in new_rules if not rr in closure])
                closure = closure | set(new_rules)
        return [(r,0) for r in closure]

    def extend_state(self, state):
        """
        Group the rule by the symbol just before the dot. Then, for each of those
        rules where we can move the dot further to the right, we create a new
        state and repeat the process. We also create a transition from the current
        state to the new state for the given terminal / non-terminal symbol that
        we found.

        For the rules that are already fully reduced, we add a 'reduce' entry,
        specifying the reduction that should take place.
        """
        j = self.states.index(state)
        rules_by_symbol = defaultdict(list)
        rules_to_reduce = []
        for r, p in state:
            if len(self.grammar[r]) > p+1:
                symbol = self.grammar[r][p+1]
                rules_by_symbol[symbol].append((r,p+1))
            else:
                rules_to_reduce.append(r)
        if rules_to_reduce:
            self.transitions[j]['__reduce__'] = rules_to_reduce
        for symbol, new_rules in rules_by_symbol.items():
            if self.debug:
                print(symbol, new_rules)
            new_state = set(new_rules)
            for rr in new_rules:
                new_state = new_state | set(self.get_closure(*rr))
            if self.debug:
                print(new_state)
            try:
                i = self.states.index(new_state)
            except ValueError:
                i = len(self.states)
                self.states.append(new_state)
                self.extend_state(new_state)
            self.transitions[j][symbol] = i
        if self.debug:
            print(rules_by_symbol)

    def generate_states_and_transitions(self):
        self.non_terminals = set([rule[0] for rule in self.grammar])
        self.states = list()
        self.transitions = defaultdict(dict)
        state_rules = []
        state_rules.append((0,0))
        state_rules.extend(self.get_closure(0, 0))
        state = set(state_rules)
        self.states.append(state)
        self.extend_state(state)

        # this is an optimization to reduce the number of transitions
        # we need to check...
        self.callable_transitions = defaultdict(dict)

        for k, v in self.transitions.items():
            for kv, vv in v.items():
                if callable(kv):
                    self.callable_transitions[k][kv] = v

        return self.states, self.transitions, self.callable_transitions

    def get_paths(self, node, depth):
        if depth == 0:
            return [[(tuple(),node)]]
        paths = []
        for sem_value,parent in node[2]:
            parent_paths = self.get_paths(parent, depth-1)
            for parent_path in parent_paths:
                paths.append([(sem_value, node)]+parent_path)
        return paths

    def rule_as_str(self, i):
        return u'{} \u2192 {}'.format(self.grammar[i][0],' '.join([str(s) for s in self.grammar[i][1:]]))

    def get_stack_head(self, stack_heads, state):
        for head in stack_heads:
            if head[0] == state:
                return head
        return None

    def shift_stack_heads(self, stack_heads_to_process, input):

        new_stack_heads = []
        while stack_heads_to_process:
            stack_head = stack_heads_to_process.pop()
            current_state, i, parents = stack_head
            if self.debug:
                print(i,"---", current_state, parents)
            if i < len(input):
                current_symbol = input[i]
            else:
                current_symbol = '$'

            if self.debug:
                print("\n\nShifting stack heads with state '{}'".format(current_symbol))

            semantic_value = None
            transition = None

            for tr in self.callable_transitions[current_state]:
                if callable(tr):
                    tokens = tr(input[i:])
                    if tokens:
                        transition = self.transitions[current_state][tr]
                        semantic_value = tokens
            if current_symbol in self.transitions[current_state]:
                transition = self.transitions[current_state][current_symbol]
                semantic_value = tuple([current_symbol])

            if self.debug:
                print(self.transitions[current_state], current_state, transition)

            if semantic_value:
                if self.debug:
                    print("Shifting:", semantic_value)
                new_stack_head = (transition, i+len(semantic_value), [(semantic_value,stack_head)])
                existing_stack_head = self.get_stack_head(new_stack_heads, new_stack_head)
                if existing_stack_head:
                    existing_stack_head[2].append((semantic_value,stack_head))
                else:
                    new_stack_heads.append(new_stack_head)
        return new_stack_heads

    def reduce_stack_heads(self, stack_heads_to_process, input):
        new_stack_heads = []
        while stack_heads_to_process:
            stack_head = stack_heads_to_process.pop()
            current_state, i, parents = stack_head
            if not stack_head in new_stack_heads and i < len(input):
                new_stack_heads.append(stack_head)
            #if self.debug:
            #    print("\n\nProcessing stack head:",stack_head)
            #print("State:", current_state, "symbol:", input[i],"i:",i)
            #we perform all possible reduce actions...
            if '__reduce__' in self.transitions[current_state]:
                reduce_rules = self.transitions[current_state]['__reduce__']
                for reduce_rule in reduce_rules:
                    non_terminal = self.grammar[reduce_rule][0]
                    if self.debug:
                        print("\nReducing with rule",self.rule_as_str(reduce_rule))
                    reduce_length = len(self.grammar[reduce_rule])-1
                    paths = self.get_paths(stack_head, reduce_length)
                    for path in paths:
                        sem_value, ancestor = path[-1]
                        semantic_value = tuple((non_terminal,tuple([k[0] for k in path[::-1][1:]])))
                        if non_terminal == 'S':#this is the end state
                            new_state = -1
                        else:
                            new_state = self.transitions[ancestor[0]][non_terminal]
                        existing_stack_head = self.get_stack_head(new_stack_heads, new_state)
                        if existing_stack_head:
                            #if self.debug:
                            #    print(existing_stack_head)
                            if not ancestor in [a[1] for a in existing_stack_head[2]]:
                                existing_stack_head[2].append((semantic_value,ancestor))
                                stack_heads_to_process.insert(0,existing_stack_head)
                            else:
                                other_ancestor = [a for a in existing_stack_head[2] if a[1] == ancestor][0]
                                other_semantic_value = other_ancestor[0]
                                if other_semantic_value != semantic_value:
                                    if self.debug:
                                        print("Competing interpretations!")
                                        #print(semantic_value,"vs.",other_semantic_value)
                        else:
                            new_stack_head = (new_state, i, [(semantic_value, ancestor)])
                            new_stack_heads.append(new_stack_head)
                            stack_heads_to_process.insert(0,new_stack_head)
        #if self.debug:
        #    print(new_stack_heads)
        return new_stack_heads

    def run(self, input):
        stack_heads =[
            (0, 0,[])
        ]
        accepted_stacks = []
        output_stream = []

        while True:

            if self.debug:
                print("{} stack heads".format(len(stack_heads)))
                print("\n".join([str(s) for s in stack_heads]))

            new_stack_heads = self.reduce_stack_heads(stack_heads, input)

            for stack_head in new_stack_heads:
                if stack_head[1] == len(input) + 1:
                    accepted_stacks.append(stack_head)

            stack_heads = self.shift_stack_heads(new_stack_heads, input)

            if not stack_heads:
                break

        return accepted_stacks

input_string = r"""

model Workshop {

  Usage {
    This model represents a workshop. It ... \}
    wer werwerjnwer werlwer lkjsdf slfjwerw elkjsddfsdcslkdjfsdfsfdj
    sdf sfdlksjd flkjs dlfkjsdf lkjscds sdflkjsdf lkjsdf sdsdflkjsdf lkj
    sdfölksd fsdfölk sdfölksdf ölksdcrefetpgkposdkfdgkdfgölk dscölksdf sfdsd
    sdf sösdf ölksdf söfsdölksd fsdfölkf sdöksdf ösldkf lksdölksdölksdfsdf
    sdföksdf sdfökdsf csdkeiwdsdfölkcd gflkdfgölkrtxcd sdfks fsdfölksdcölscd
    sdfsdf ierimdvcsdf äölk sdfölks dfsdflksdfklscsd sdölsdfölkcsdsdc
  }

  title string {

    Usage {
        String can be used as a ....
        sd lksdf ölsdf ölksdf ölksdfölksdfsf
    }

    Constraints {
        
    }
  }

  sections []WorkshopSection
}

model WorkshopSection {
  number int
  section Section
}

enum SectionType { PLAIN, CODE, DOCKER_COMPOSE }

struct Section {
  type SectionType
  content blob
}
"""

if __name__ == '__main__':
    parser = Parser(gospel_grammar, debug=False)
    import pprint
    print("States:")
    pprint.pprint([list(state) for state in parser.states])
    print("Transitions:")
    pprint.pprint(dict(parser.transitions))
    for i in range(len(parser.grammar)):
        print(i,":",parser.rule_as_str(i))

    input_string = ['n','+','n','+','n','$']
    input_string = ['(',')']
    input_string = ['symbol','times','symbol','times','number', '+','symbol']
    input_string = ['a','a','a','a','a','a','a']
    input_string = ['b','+','b']
    input_string = ['def','name','(','param',',','name',',','param',',','name','=','"','characters',
                    '"', ',','keyword_param',',', ')',':','newline','indent','pass','newline','pass','dedent','newline',
                    'name','=','number','newline','number', 'newline','funcdef']
    input_string = gospel_program
    #input_string = ['a','n','n','a']
    #input_string = ['def','name','(',')',':','newline','indent','pass','dedent','newline']
    #input_string = ['statement','newline', 'statement','newline', 'statement','newline', 'statement','newline']
    # for i in range(18):
    #    input_string = input_string+['+']+input_string[:3]
    # print(len(input_string))
    # input_string ='1+2+1+121'
    start = time.time()
    n = 100
    print(" ".join(input_string))
    for i in range(n):
        stack_heads = parser.run(input_string)
    stop = time.time()
    print(1.0/(len(input_string)/((stop-start)/n))*1000.0)
    print("Accepted stacks: {}".format(len(stack_heads)))
    for stack_head in stack_heads:
        for semantic_value, parent in stack_head[2]:
            pprint.pprint(semantic_value)
    exit(0)
