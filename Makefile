.PHONY: test

test:
	$(MAKE) -C core test
lint:
	$(MAKE) -C core lint