
all:
	@(cd build/node; $(MAKE))
	@(cd build/parser_generator; $(MAKE))
	@(cd build/stringparser; $(MAKE))

install:
	@(cd build/node; $(MAKE) install)
	@(cd build/parser_generator; $(MAKE) install)
	@(cd build/stringparser; $(MAKE) install)

clean:
	@(cd build/node; $(MAKE) clean)
	@(cd build/parser_generator; $(MAKE) clean)
	@(cd build/stringparser; $(MAKE) clean)
