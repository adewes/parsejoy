
(1,2,3,4,5)


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
        self.initialized = True

        def add_token_list(token_list,parent,next_parent,previous_parent,token_id=0):
            previous_token = previous_parent
            wrapped_tokens = []
            for i,token in enumerate(token_list):
                wrapped_tokens.append({
                    'value' : token, \
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


def my_decorator(boobo : "dsfs",baaa, baz : test = "foo", *args, **kwargs):
    pass

@my_decorator(baz ="bar")
class Foo(bar):
    pass
    """
    this is my docstring
    """

    #this will break everything
if False is not True:
    pass
    pass #this should work...  
    pass#this as well
    a = b+c + d \
       + es -dsfs \
       + sdf
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
else:
    fopo
    def foo(nar, bar, zar):
        pass
        if a == b:
            pass
    pass ##sfsdfdsfdsf
pass


wow = "test"
##and this
foo = """

\"""

""\"

sdfsfs"""

bar = '''
\'''
'\''
''\'

sdfsfsd f
ds
fdsfsd
fsd
fds
fdsfsd

        dssfdsgfsfdöäö
        sdfsfsd
                sdfsdfdsf   sdfdsfsd
'''

pass

fooz

def decorated_function(aa,*a,**aasfdsf):
    pass





def my_decorator(boobo : "dsfs",baaa, baz : test = "foo", *args, **kwargs):
    pass

@my_decorator(baz ="bar")
class Foo(bar):
    pass
    """
    this is my docstring
    """

    #this will break everything
if False is not True:
    pass
    pass #this should work...  
    pass#this as well
    a = b+c + d \
       + es -dsfs \
       + sdf
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
    c = ["a","b","d",  "e",  "f","g","h","i","j","k","l"]
else:
    fopo
    def foo(nar, bar, zar):
        pass
        if a == b:
            pass
    pass ##sfsdfdsfdsf
pass


wow = "test"
##and this
foo = """

\"""

""\"

sdfsfs"""

bar = '''
\'''
'\''
''\'

sdfsfsd f
ds
fdsfsd
fsd
fds
fdsfsd

        dssfdsgfsfdöäö
        sdfsfsd
                sdfsdfdsf   sdfdsfsd
'''

pass

fooz

def decorated_function(aa,*a,**aasfdsf):
    pass
    
