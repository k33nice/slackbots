include .env
SHELL = bash

LINUX := linux/amd64
OSX := darwin/amd64
PLATFORMS := $(LINUX) $(OSX)

HOST ?= "127.0.0.1"
NAME ?= "name"

# reverse list of words e.g. "foo baz bar" => "bar baz foo"
reverse = $(if $(1),$(call reverse,$(wordlist 2,$(words $(1)),$(1)))) $(firstword $(1))

last_word = $(words $(temp))
temp = $(subst /, ,$@)
os = $(word 2, $(call reverse,$(temp)))
arch = $(word $(last_word), $(temp))

.PHONY: test, bench, build, rollout, clean

clean:
	rm build/*

test:
	go test -v .

bench:
	go test -bench .

build: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -o build/'$(os)-$(arch)'

rollout: $(LINUX) rollout/$(LINUX)

$(addprefix rollout/, $(PLATFORMS)):
	scp build/$(os)-$(arch) root@$(HOST):/usr/sbin/$(NAME)
	ssh root@$(HOST) mkdir -p /etc/slackbot/
	ssh root@$(HOST) mkdir -p /var/log/slackbot/
	scp etc/config.toml root@$(HOST):/etc/slackbot/
	scp misc/$(NAME).service root@$(HOST):/etc/systemd/system/
