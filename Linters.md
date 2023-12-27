Linters and other analysis tools

- [golangci-lint](#golangci-lint)
- [gofumpt](#gofumpt)
- [wsl](#wsl)
- [golines](#golines)
- [wrapmsg](#wrapmsg)
- [lingo](#lingo)
- [go-ssaviz](#go-ssaviz)
- [skeleton](#skeleton)
- [ssadump](#ssadump)
- [astree](#astree)
- [deferror](#deferror)

## [golangci-lint](https://github.com/golangci/golangci-lint)

```bash
golangci-lint --enable-all -v run --disable depguard,varnamelen,gomnd,nlreturn,exhaustivestruct,exhaustruct,nonamedreturns
```

## [gofumpt](https://github.com/mvdan/gofumpt)

Stricter gofmt

```bash
gofumpt -w fileName
```

## [wsl](https://github.com/bombsimon/wsl)

Add whitespaces

```bash
wsl -fix fileName
```

## [golines](https://github.com/segmentio/golines)

Reformat the code when it exceeds a certain column length

```bash
# 80 columns length
golines -w -m 80 fileName
```

## [wrapmsg](https://github.com/Warashi/wrapmsg)

Checks the Wrap error message

```bash
go vet -vettool=$(which wrapmsg) ./...
```

## [lingo](https://github.com/sgatev/lingo)

Checks and enforces Go lingo

```yaml
# lingo.yml
matchers:
  - type: "glob"
    config:
      pattern: "**/*.go"
  - type: "not"
    config:
      type: "glob"
      config:
        pattern: "**/*_test.go"

checkers:
  local_return:
  multi_word_ident_name:
  exported_ident_doc:
  consistent_receiver_names:
  left_quantifiers:
  pass_context_first:
  return_error_last:
```

```bash
lingo check ./...
```

## [go-ssaviz](https://github.com/SilverRainZ/go-ssaviz)

Very cool but I really have no clue how to make use of it.

```bash
go-ssaviz ./...
```

## [skeleton](https://github.com/gostaticanalysis/skeleton)

Static analisys tool code generator.

## [ssadump](https://github.com/golang/tools/blob/master/cmd/ssadump/main.go)

```bash
go install golang.org/x/tools/cmd/ssadump@latest

ssadump -build=F main.go
```

## [astree](https://github.com/knsh14/astree)

```bash
go install github.com/knsh14/astree/cmd/astree@latest

astree main.go
```

## [deferror](https://github.com/Jiang-Gianni/deferror)

Personal made linter to add a defer function call when there is a named return `err`.

Not for Go code but: [kube-score](https://github.com/zegl/kube-score) Kubernetes, [actionlint](https://github.com/rhysd/actionlint) Github Actions,
