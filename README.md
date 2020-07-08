# Accio
Accio is a flexible framework for boilerplate code generators. It is designed with readability in mind because logic-full templates are hard to maintain. Its modular approach to templates makes it easy and fun to work with, and the possibility to script - just powerful enough to handle most edge-cases.

## Features
* Prompts for data input
* Scripting support with Starlark (Python dialect)
* Running generators locally or directly from git repositories
* Unique markup language designed for complex text
* No dependencies - just a single binary
* Cross-platform (Linux, OS X and Windows)

## Installation
### Pre-compiled binaries
*Homebrew*:

`brew install g1ntas/tap/accio`

*Shell script*:

`curl -sfL https://raw.githubusercontent.com/g1ntas/accio/master/install.sh | sh`

*Manually*:

 Download the pre-compiled binaries from [releases page](https://github.com/g1ntas/accio/releases).

### Building from source
To build a binary from the source code, you need to have [Go](https://golang.org/) installed first.

*Steps*:
1. Clone repository: `git clone https://github.com/g1ntas/accio`
2. Build: `go build`
3. Run `./accio` to verify if it works

*NOTE*: 
While it’s the easiest way to build a binary, it doesn’t include any versioning information. For best results, use [Mage](https://magefile.org/) with `mage build`.

## Quickstart
### Running generator
You can run a generator from a local directory with `run` command:

`accio run ./generator-directory`

Or from any Git repository (and subdirectories are supported!):

`accio run github.com/g1ntas/accio/examples/open-source-license`

### Creating first generator
Create a config file `~/example/.accio.toml`:
```toml
# Define a prompt, which will be shown when the generator is executed
[prompts.filename]
type="input"
message="Enter filename:"
```

Create a template file `~/example/file.accio`:
```
# Make our entered filename uppercase with Starlark code 
variable -name="uppercaseFilename" <<
    return vars['filename'].upper()
>>

# Rename the file
filename << 
    return vars['uppercaseFilename']
>>

# Use mustache templating engine to output content of the file
template <<
Name of this file is: {{uppercaseFilename}}
>>
```

Run the generator:
```
> accio run ~/example
$ Enter filename:
> test.txt
$ [SUCCESS] ~/example/TEST.TXT created.
> cat ~/example/TEST.TXT
$ Name of this file is: TEST.TXT
```

## Examples
* [github.com/g1ntas/accio/examples/go-travisci-config](examples/go-travisci-config) - generates TravisCI config with selected Go versions, operating systems, and CPU architectures
* [github.com/g1ntas/accio/examples/golang-cli-project](examples/golang-cli-project) - generates a boilerplate Go project with a Cobra CLI command
* [github.com/g1ntas/accio/examples/open-source-license](examples/open-source-license) - generates the selected open-source license

## Documentation
* Core concepts
	* [Generators](docs/concepts/generators.md)
	* [Templates](docs/concepts/templates.md)
		* [Static templates](docs/concepts/templates.md#static-templates)
		* [Blueprints](docs/concepts/templates.md#blueprints)
* Reference
	* [Configuration](docs/reference/configuration.md)
		* [help](docs/reference/configuration.md#help)
		* [prompts](docs/reference/configuration.md#prompts)
	* [Blueprints](docs/reference/blueprints.md)
		* [Starlark](docs/reference/blueprints.md#starlark)
		* [Mustache](docs/reference/blueprints.md#mustache)
		* [Tags](docs/reference/blueprints.md#tags)
			* [filename](docs/reference/blueprints.md#filename)
			* [variable](docs/reference/blueprints.md#variable)
			* [skipif](docs/reference/blueprints.md#skipif)
			* [partial](docs/reference/blueprints.md#partial)
			* [template](docs/reference/blueprints.md#template)
	* [Accio markup language](docs/reference/accio-ml.md)
		* [Comments](docs/reference/accio-ml.md#comments)
		* [Identifiers](docs/reference/accio-ml.md#identifiers)
		* [Tags](docs/reference/accio-ml.md#tags)
			* [Attributes](docs/reference/accio-ml.md#attributes)
			* [Body](docs/reference/accio-ml.md#body)
			* [Inline body](docs/reference/accio-ml.md#inline-body)
		* [Custom delimiters](docs/reference/accio-ml.md#custom-delimiters)

## Contributing
Contributions are more than welcome, if you are interested, feel free to open an issue or create a pull request.

## Copyright
Accio is released under the MIT license. See [LICENSE](LICENSE).