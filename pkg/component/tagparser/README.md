A parser for conventional Go struct tags
========================================

[![Go reference](https://pkg.go.dev/badge/github.com/talav/talav/pkg/component/tagparser.svg)](https://pkg.go.dev/github.com/talav/talav/pkg/component/tagparser) ![Zero dependencies](https://img.shields.io/badge/deps-zero-brightgreen) ![Zero magic](https://img.shields.io/badge/magic-none-brightgreen)

Parses the conventional format of struct field tags: `name,key1,key2=value2,key3='value with spaces, equals signs, and \' quotes',key4`.

Automatically handles Go struct tag quoting conventions (e.g., `"name=value"` → `name=value`).

This parser enforces strict quoting rules and provides comprehensive error reporting with position information.


Usage
-----

Use `Parse` to parse tags that treat the first item as a name:

```go
tag, err := tagparser.Parse(`foo,bar,boz='buzz fubar'`)
// tag.Name == "foo"
// tag.Options == map[string]string{"bar": "", "boz": "buzz fubar"}

// If the first item has an equals sign, it's treated as a key-value pair:
tag2, _ := tagparser.Parse(`foo=bar,baz`)
// tag2.Name == ""
// tag2.Options == map[string]string{"foo": "bar", "baz": ""}
```

Use `ParseFunc` for customized parsing and zero allocations:

```go
var name string
opts := make(map[string]string)

err := tagparser.ParseFunc(`foo,bar=xx,bar=yy`, func(key, value string) error {
    // Empty key means this is the first item (name)
    if key == "" {
        name = value
        return nil
    }
    // Last value wins for duplicates
    opts[key] = value
    return nil
})
// name == "foo"
// opts == map[string]string{"bar": "yy"}
```

Empty values are allowed:

```go
tag, err := tagparser.Parse(`foo,bar=`)
// tag.Name == "foo"
// tag.Options == map[string]string{"bar": ""}
```


Error handling
--------------

All errors returned are `*tagparser.Error`, providing a clear message and a string index of the error position. The error content is not covered by compatibility guarantees.

Note that you can simply ignore errors if you like; the parser still returns the best guess about the meaning of the tag.


Tag syntax
----------

* A tag is a list of comma-separated items.

* An item is either a `key=value` pair or just a single string.

* Both keys and values can be bare words (`foo= bar`) or single-quoted strings (`foo= 'bar= boz, buzz and fubar'`). Quotes, if present, must enclose the entire value after trimming whitespace. Mixed quoting like `foo'bar'` is not allowed.

* Both keys and values can use a backslash to escape special characters (`foo\ bar`, `foo\=bar`, `foo\,bar`, `'foo\'n\'bar'`). In bare strings, escape equals signs and commas. In quoted strings, escape quotes and backslashes. Examples:
  - Bare: `foo\=bar` → "foo=bar"
  - Quoted: `'foo\'bar'` → "foo'bar"
  - Quoted: `'foo\\bar'` → "foo\bar"
  
  The escapes are processed and removed from the values (so `foo=\=\,\!` is returned as `map[string]string{"foo": "=,!"}`); you can escape any non-alphabetical characters.

* Non-escaped unquoted leading and trailing ASCII whitespace is trimmed from keys and values. Escaped whitespace is preserved (e.g., `\ ` remains as a space character).

* `Parse` and `ParseFunc` give special treatment to the first item of the tag if it does not have an equals sign. Such an item is returned as `Tag.Name` by `Parse` / as a value with an empty key by `ParseFunc`. If the first item does have an equals sign, it is treated as a normal key; `Parse` returns an empty `Tag.Name`, and `ParseFunc` reports a normal item and does not report an item with an empty key.

* Duplicate keys are allowed; the last value wins (e.g., `key=first,key=second` results in `key=second`).

* For normal items, empty key names are not allowed. Empty values are allowed (e.g., `key=` is valid and represents an empty string value).