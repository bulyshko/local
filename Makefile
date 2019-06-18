NAME=local
VERSION?=latest

.PHONY: build save clean

build: Dockerfile
	docker build -t $(NAME):$(VERSION) .

save: build
	docker save -o $(NAME)-$(VERSION).tar $(NAME):$(VERSION)

clean:
	rm -f $(NAME)-$(VERSION).tar
