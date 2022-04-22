YALIVE_VERSION = $(shell cat version.txt)

BINARY ?= yalive-server

# test:
# 	go test ./...

local-build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/$(BINARY) .

check-go-version:
	@go version | grep ' go1\.18\.1 ' || (echo 'Please install Go version 1.18.1' && false)

version-go:
	node scripts/build.js --update-version-go

platform-all:
	@$(MAKE) --no-print-directory -j4 \
		platform-darwin \
		platform-darwin-arm64 \
		platform-linux \
		platform-linux-arm64 \
		platform-windows \
		platform-windows-arm64 \
		platform-neutral

platform-windows: version-go
	node scripts/build.js npm/yalive-server-windows-64/package.json --version
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GO_FLAGS) -o npm/yalive-server-windows-64/bin/yalive-server.exe .

platform-windows-arm64: version-go
	node scripts/build.js npm/yalive-server-windows-arm64/package.json --version
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(GO_FLAGS) -o npm/yalive-server-windows-arm64/bin/yalive-server.exe .

platform-unixlike: cmd/version.go
	@test -n "$(GOOS)" || (echo "The environment variable GOOS must be provided" && false)
	@test -n "$(GOARCH)" || (echo "The environment variable GOARCH must be provided" && false)
	@test -n "$(NPMDIR)" || (echo "The environment variable NPMDIR must be provided" && false)
	node scripts/build.js "$(NPMDIR)/package.json" --version
	CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build $(GO_FLAGS) -o "$(NPMDIR)/bin/yalive-server" .

platform-darwin:
	@$(MAKE) --no-print-directory GOOS=darwin GOARCH=amd64 NPMDIR=npm/yalive-server-darwin-64 platform-unixlike

platform-darwin-arm64:
	@$(MAKE) --no-print-directory GOOS=darwin GOARCH=arm64 NPMDIR=npm/yalive-server-darwin-arm64 platform-unixlike

platform-linux:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=amd64 NPMDIR=npm/yalive-server-linux-64 platform-unixlike

platform-linux-arm64:
	@$(MAKE) --no-print-directory GOOS=linux GOARCH=arm64 NPMDIR=npm/yalive-server-linux-arm64 platform-unixlike

platform-neutral:
	node scripts/build.js npm/yalive-server/package.json --version
	node scripts/build.js --neutral

# slightly modified
prepare-publish: check-go-version
	@pnpm --version > /dev/null || (echo "The 'pnpm' command must be in your path to publish" && false)
	@echo "Checking for uncommitted/untracked changes..." && test -z "`git status --porcelain | grep -vE 'M (CHANGELOG\.md|version\.txt)'`" || \
		(echo "Refusing to publish with these uncommitted/untracked changes:" && \
		git status --porcelain | grep -vE 'M (CHANGELOG\.md|version\.txt)' && false)
	@echo "Checking for main branch..." && test main = "`git rev-parse --abbrev-ref HEAD`" || \
		(echo "Refusing to publish from non-main branch `git rev-parse --abbrev-ref HEAD`" && false)
	@echo "Checking for unpushed commits..." && git fetch
	@test "" = "`git cherry`" || (echo "Refusing to publish with unpushed commits" && false)

	# Prebuild now to prime go's compile cache and avoid timing issues later
	@$(MAKE) --no-print-directory platform-all

	# Commit now before publishing so git is clean for this: https://github.com/golang/go/issues/37475
	# Note: If this fails, then the version number was likely not incremented before running this command
	git commit -am "publish $(YALIVE_VERSION) to npm"
	git tag "v$(YALIVE_VERSION)"
	@test -z "`git status --porcelain`" || (echo "Aborting because git is somehow unclean after a commit" && false)

	# Make sure the npm directory is pristine (including .gitignored files) since it will be published
	# rm -fr npm && git checkout npm

publish-all: check-go-version
	@echo Enter one-time password:
	@read OTP && OTP="$$OTP" $(MAKE) --no-print-directory -j4 \
		publish-windows \
		publish-windows-arm64

	@echo Enter one-time password:
	@read OTP && OTP="$$OTP" $(MAKE) --no-print-directory -j4 \
		publish-darwin \
		publish-darwin-arm64

	@echo Enter one-time password:
	@read OTP && OTP="$$OTP" $(MAKE) --no-print-directory -j4 \
		publish-linux \
		publish-linux-arm64

	# Do these last to avoid race conditions
	@echo Enter one-time password:
	@read OTP && OTP="$$OTP" $(MAKE) --no-print-directory -j4 \
		publish-neutral

	git push origin main "v$(YALIVE_VERSION)"

publish-windows: platform-windows
	test -n "$(OTP)" && cd npm/yalive-server-windows-64 && pnpm publish --otp="$(OTP)"

publish-windows-arm64: platform-windows-arm64
	test -n "$(OTP)" && cd npm/yalive-server-windows-arm64 && pnpm publish --otp="$(OTP)"

publish-darwin: platform-darwin
	test -n "$(OTP)" && cd npm/yalive-server-darwin-64 && pnpm publish --otp="$(OTP)"

publish-darwin-arm64: platform-darwin-arm64
	test -n "$(OTP)" && cd npm/yalive-server-darwin-arm64 && pnpm publish --otp="$(OTP)"

publish-linux: platform-linux
	test -n "$(OTP)" && cd npm/yalive-server-linux-64 && pnpm publish --otp="$(OTP)"

publish-linux-arm64: platform-linux-arm64
	test -n "$(OTP)" && cd npm/yalive-server-linux-arm64 && pnpm publish --otp="$(OTP)"

publish-neutral: platform-neutral
	test -n "$(OTP)" && cd npm/yalive-server && pnpm publish --otp="$(OTP)"
