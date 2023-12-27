Go concepts and syntax

- [**Integer literals** *(100 Go Mistakes #17)*](#integer-literals-100-go-mistakes-17)
- [**Slice Copy** *(100 Go Mistakes #25)*](#slice-copy-100-go-mistakes-25)
- [**Map initalization** *(100 Go Mistakes #27)*](#map-initalization-100-go-mistakes-27)
- [**Break** *(100 Go Mistakes #34)*](#break-100-go-mistakes-34)
- [**Rune** *(100 Go Mistakes #36 and #37)*](#rune-100-go-mistakes-36-and-37)
- [**JSON Marshal and Unmarshal** *(100 Go Mistakes #77)*](#json-marshal-and-unmarshal-100-go-mistakes-77)
- [**GoString**](#gostring)
- [Interface compile time check](#interface-compile-time-check)
- [**Reflect to assign value**](#reflect-to-assign-value)
- [**Generics Constraints** *(100 Go Mistakes #9)*](#generics-constraints-100-go-mistakes-9)


## **Integer literals** *(100 Go Mistakes #17)*

| Literal     | Prefixes        |
| ----------- | --------------- |
| Binary      | 0b, 0B          |
| Octal       | 0,       0o, 0O |
| Hexadecimal | 0x, 0X          |

```go
// Output: 102
fmt.Println(100 + 0b10)

// Output: 108
fmt.Println(100 + 010)

// Output: 108
fmt.Println(100 + 0o10)

// Output: 116
fmt.Println(100 + 0x10)
```


## **Slice Copy** *(100 Go Mistakes #25)*

`copy` copies the minimum number of elements between the length of the source and the length of the destination

An alternative is `append([]int{}, src...)`

```go
	src := []int{0, 1, 2}
	var dst1 []int
	copy(dst1, src)
	// []
	fmt.Println(dst1)
	var dst2 = make([]int, len(src))
	copy(dst2, src)
	// [0, 1, 2]
	fmt.Println(dst2)
	var dst3 = append([]int{}, src...)
	// [0, 1, 2]
	fmt.Println(dst3)
```


## **Map initalization** *(100 Go Mistakes #27)*

Like for slices, an insertion can be an O(n) operation.

If the maximum possible length is known, then use `make` to initialize the map.

```go
m := make(map[int]int, len(valuesToStore))
```


## **Break** *(100 Go Mistakes #34)*

`break` terminates the execution of the innermost `for`, `switch` or `select` statement.

Use a label to specify which statement to stop.

`continue` works similarly.

```go
fast:
	for i := 0; i < 5; i++ {
		fmt.Println(i)
		switch i {
		default:
		case 2:
			break fast
		}
	}
```


## **Rune** *(100 Go Mistakes #36 and #37)*

* charset = set of characters
* encoding = how to translate a charset to binary (in UTF-8: 汉 => []byte{0xE6, 0xB1, 0x89}, between 1 and 4 bytes)
  ```go
  	s := string([]byte{0xE6, 0xB1, 0x89})
	fmt.Printf("%s\n", s) // Prints 汉
  ```
* code point = single value (汉 => U+6C49)
* len(myString) returns the number of bytes
  ```go
  	s := "汉"
	fmt.Println(len(s)) // Prints 3
  ```
* rune = int32 = code point

Example:

```go
	s := "hêllo"
	for i, r := range s {
		fmt.Printf("position %d: %c, %c\n", i, s[i], r)
	}
	fmt.Printf("len=%d\n", len(s))
	// Output
	// position 0: h, h
	// position 1: Ã, ê
	// position 3: l, l
	// position 4: l, l
	// position 5: o, o
	// len=6
```

* ê requires 2 bytes (len = 6)
* range iterates over the start of each rune (i = 2 is skipped)
* s[i] prints the UTF-8 representation of the byte


## **JSON Marshal and Unmarshal** *(100 Go Mistakes #77)*

```go
type Event struct {
	ID int
	time.Time
}
```

`time.Time` implements the `Marshaler` interface (which requires a `MarshalJSON() ([]byte, error)` function)

`json.Marshal(event)` will ignore the `ID int` field and only return the time because the embedded `time.Time`'s `MarshalJSON` method has been promoted (the default behavior is ignored).

Either assign a field name to `time.Time` (no embedded struct) or define the `MarshalJSON` implementation of `Event`.

`time.Time` has *wall* (time of day) and *monotonic* (moves only forward) clocks. `json.Unmarshal` a `time.Time` field only returns the *wall* clock.

The `time.Equal` function will ignore the *monotonic* clock

```go
event1.Time.Equal(event2.Time)
```

The `time.Truncate(0)` function will strip away the *monotonic* clock

```go
time.Now().Truncate(0)
```


## [**GoString**](https://www.youtube.com/watch?v=HTIltI0NuNg&t=21m30s)

`%#v` will format the struct into a string with the name of the fields

```go
package main

import "fmt"

type Record struct {
	MyField int
}

func main() {
	r := &Record{MyField: 1234}
	fmt.Printf("%#v\n", r)
	// Prints out:
	// &main.Record{MyField:1234}
}
```

It is possible to define a custom format by defining a `Gostring` method on the struct

```go
func (r *Record) GoString() string {
	return fmt.Sprintf("MyField is %d", r.MyField)
}
```

This is useful in case, for example, one of the field is a pointer and you want to print out the actual value instead of the memory address.

This can also make it easier to copy paste it to assign it to a variable in case of debugging.

The following prints `&main.Record{MyField: func(v int) *int { return &v }(1234)}` instead of something like `&main.Record{MyField:(*int)(0xc00013e090)}`.

```go
type Record struct {
	MyField *int
}

func main() {
	myInt := 1234
	r := &Record{MyField: &myInt}
	fmt.Printf("%#v\n", r)
}

func (r *Record) GoString() string {
	return fmt.Sprintf(
		"&main.Record{MyField: func(v int) *int { return &v }(%v)}",
		*r.MyField,
	)
}
```

Can be auto generated by `https://github.com/awalterschulze/goderive`


## [Interface compile time check](https://go.dev/doc/effective_go#blank_implements)

Use a blank identifier to ensure that a struct type (`RawMessage`) implements a interface (`json.Marshaler`)

```go
var _ json.Marshaler = (*RawMessage)(nil)
```


## [**Reflect to assign value**](https://www.youtube.com/watch?v=hz6d7rzqJ6Q&t=6m15s)

It is possible to scan a struct and replace the value of the fields based on struct tags ([11m 20s](https://www.youtube.com/watch?v=hz6d7rzqJ6Q&t=6m15s))


## **Generics Constraints** *(100 Go Mistakes #9)*

```go
type customConstraint interface {
	~int | ~string
}

func getKeys[K customConstraint, V any](m map[K]V) []V {
	keys := make([]V, len(m))
	i := 0
	for k := range m {
		keys[i] = m[k]
		i++
	}
	return keys
}
```

* | is the union operator
* ~int and ~string include all the types whose underlying type is an int or a string
