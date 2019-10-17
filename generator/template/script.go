package template

import (
	"fmt"
	"go.starlark.net/starlark"
	"log"
)

type script struct {
	body string
	inline bool
}

func (t *script) execute(vars map[string]string) {
	const data = `
print(greeting + ", world")
squares = [x*x for x in range(10)]
`

	// The Thread defines the behavior of the built-in 'print' function.
	thread := &starlark.Thread{
		Name:  "example",
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
	}

	// This dictionary defines the pre-declared environment.
	predeclared := starlark.StringDict{
		"greeting": starlark.String("hello"),
		//"repeat":   starlark.NewBuiltin("repeat", repeat),
	}

	// Execute a program.
	globals, err := starlark.ExecFile(thread, "apparent/filename.star", data, predeclared)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}

	// Print the global environment.
	fmt.Println("\nGlobals:")
	for _, name := range globals.Keys() {
		v := globals[name]
		fmt.Printf("%s (%s) = %s\n", name, v.Type(), v.String())
	}
}

func execute(content string) {

}