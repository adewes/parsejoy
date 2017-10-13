SRCDIR=src
BUILDDIR=build
C_BUILDDIR=build/c
PYTHON_BUILDDIR=build/python
DEPSDIR=dependencies

YAML_CPP_VERSION=0.5.3
LUAJIT_VERSION=2.0.4
BOOST_VERSION=1.61.0
LIBCRYPTO_VERSION=5.6.4

SWIG_OPTS=
PYTHON_VERSION=2.7

SRCFILES = set.cc node.cc parser.cc stringparser.cc grammar.cc lua_environment.cc parsejoy_wrap_lua.cc
HEADERFILES=$(addprefix $(SRCDIR)/,$(notdir $(SRCFILES:.cc=.h)))
OBJFILES=$(addprefix $(BUILDDIR)/,$(notdir $(SRCFILES:.cc=.o)))

SWIG_LUA_SRCDIR=src/swig/lua
SWIG_LUA_INTERFACE_FILES = $(addprefix $(SWIG_LUA_SRCDIR)/, parsejoy.i parser.i stringparser.i lua_environment.i grammar.i shared_ptr.i set.i shared_ptr.i)
SWIG_LUA_RUNTIME = $(SRCDIR)/lua_runtime.h
SWIG_LUA_SRCFILES = parsejoy_wrap_lua.cc
SWIG_LUA_OBJFILES = $(addprefix $(BUILDDIR)/,$(notdir $(SWIG_LUA_SRCFILES:.cc=.o)))

SWIG_PYTHON_SRCDIR=src/swig/python
SWIG_PYTHON_INTERFACE_FILES = $(addprefix $(SWIG_PYTHON_SRCDIR)/, parsejoy.i grammar.i)
SWIG_PYTHON_RUNTIME = $(SRCDIR)/python_runtime.h
SWIG_PYTHON_SRCFILES = parsejoy_wrap_python.cc
SWIG_PYTHON_OBJFILES = $(addprefix $(BUILDDIR)/,$(notdir $(SWIG_PYTHON_SRCFILES:.cc=.o)))

C_SRCFILES = prompt.c
C_OBJFILES=$(addprefix $(BUILDDIR)/c/,$(notdir $(C_SRCFILES:.c=.o)))
GCC_OPTS = -D HAVE_LIBREADLINE=1 -D HAVE_READLINE_HISTORY=1
GCC = gcc $(GCC_OPTS)

INCLUDEDIRS=-isystem$(SRCDIR) \
		    -isystem/usr/include/python$(PYTHON_VERSION) \
			-isystem$(DEPSDIR)/libcrypto++$(LIBCRYPTO_VERSION) \
			-isystem$(DEPSDIR)/boost-$(BOOST_VERSION) \
			-isystem$(DEPSDIR)/yaml-cpp-$(YAML_CPP_VERSION)/include \
			-isystem$(DEPSDIR)/luajit-$(LUAJIT_VERSION)/src \
			-isystem$(DEPSDIR)/re2-2016-10-01

LIBDIRS=-L$(DEPSDIR)/boost-$(BOOST_VERSION) \
		-L$(DEPSDIR)/libcrypto++$(LIBCRYPTO_VERSION) \
		-L$(DEPSDIR)/yaml-cpp-$(YAML_CPP_VERSION) \
		-L$(DEPSDIR)/luajit-$(LUAJIT_VERSION)/src \
		-L$(DEPSDIR)/re2-2016-10-01/obj \
		-L/lib/x86_64-linux-gnu/

LIBS=-lcryptopp -lluajit -lyaml-cpp 
LIBS_TOOL=-lreadline

PROFILE_OPTS =
CPP = g++ -fPIC -std=c++14 $(PROFILE_OPTS)

all: $(BUILDDIR) $(C_BUILDDIR) $(BUILDDIR)/libparsejoy.a

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

$(C_BUILDDIR):
	mkdir -p $(C_BUILDDIR)

$(PYTHON_BUILDDIR):
	mkdir -p $(PYTHON_BUILDDIR)

$(SWIG_LUA_RUNTIME):
	swig3.0 $(SWIG_OPTS) -lua -external-runtime $(SWIG_LUA_RUNTIME)

$(SWIG_PYTHON_RUNTIME):
	swig3.0 $(SWIG_OPTS) -python -external-runtime $(SWIG_PYTHON_RUNTIME)

$(BUILDDIR)/c/%.o: $(SRCDIR)/c/%.c $(C_BUILDDIR)
	$(GCC) $(INCLUDEDIRS) $(LIBDIRS) -c $< -o $@

$(BUILDDIR)/%.o: $(SRCDIR)/%.cc
	$(CPP) $(INCLUDEDIRS) $(LIBDIRS) -c $< -o $@

$(BUILDDIR)/libparsejoy.a: $(OBJFILES) $(C_OBJFILES) $(SWIG_LUA_OBJFILES)
	ar -crs $(BUILDDIR)/libparsejoy.a $(C_OBJFILES) $(OBJFILES) $(SWIG_LUA_OBJFILES)

.PHONY: all

tool: $(BUILDDIR) $(BUILDDIR)/libparsejoy.a
	$(CPP) -L$(BUILDDIR) $(LIBDIRS) $(INCLUDEDIRS) $(SRCDIR)/tool.cc -lparsejoy $(LIBS) $(LIBS_TOOL) -o $(BUILDDIR)/tool

python3lib: SWIG_OPTS=-py3
python3lib: PYTHON_VERSION=3.5
python3lib: python2lib

python2lib: $(PYTHON_BUILDDIR) $(PYTHON_BUILDDIR)/parsejoy.py $(BUILDDIR)/libparsejoy.a $(SRCDIR)/parsejoy_wrap_python.cc
	$(CPP) -shared -L$(BUILDDIR) $(LIBDIRS) $(INCLUDEDIRS) $(SRCDIR)/parsejoy_wrap_python.cc -lgcc_s -lparsejoy $(shell python$(PYTHON_VERSION)-config --libs) $(LIBS) -o $(BUILDDIR)/python/_parsejoy.so

$(SRCDIR)/parsejoy_wrap_lua.cc: $(SWIG_LUA_INTERFACE_FILES)
	swig3.0 $(SWIG_OPTS) -c++ -lua -I$(SRCDIR) -I$(SWIG_LUA_SRCDIR) -o $@ $<

$(PYTHON_BUILDDIR)/parsejoy.py $(SRCDIR)/parsejoy_wrap_python.cc: $(SWIG_PYTHON_INTERFACE_FILES)
	swig3.0 $(SWIG_OPTS) -c++ -python -I$(SRCDIR) -I$(SWIG_PYTHON_SRCDIR) -outdir $(PYTHON_BUILDDIR) -o $(SRCDIR)/parsejoy_wrap_python.cc $<


default:
	all

clean:
	rm -rf $(BUILDDIR)
	#rm -rf $(SRCDIR)/*wrap_.cc
