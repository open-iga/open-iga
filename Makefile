.PHONY: test lint

test:
	$(MAKE) -C core test
lint:
	$(MAKE) -C core lint