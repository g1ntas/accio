# Accio markup language
AccioML is a simple, flat markup language with the design goal to make it easy to work with complex text (e.g. programming code), without applying verbose escaping techniques and still keeping text easily readable by humans. The syntax of the language was inspired by [PXSL (Parsimonious XML Shorthand Language)](http://web.archive.org/web/20060228113459/http://community.moertel.com/pxsl/).

## Syntax

### Comments
Octothorpe/hash sign (`#`) appearing at the start of a new line indicates the beginning of the comment. The comment extends until the end of the line.  

Example:
```
# Comment 1
# Comment 2
```

### Identifiers
An identifier is a sequence of lexical tokens that name language entities like tags or attributes. Identifiers can consist of ASCII letters, decimal digits, or hyphens (`-`), but must not begin with a digit or hyphen and must not end with a hyphen. Identifiers are case sensitive.

### Tags
An identifier appearing on the new line indicates the start of a tag. It can contain multiple attributes and a body. A tag extends until the end of the line or until the end of the body if there is one specified.

Example:
```
empty-tag
another-tag
tag3
```

#### Attributes
Attributes appear after a tag identifier followed by at least a single space or tab character. An attribute starts with a hyphen (`-`), followed by an identifier indicating the attribute's name, followed by an equal sign (`=`), and finished by a value within the double-quotes. Value can contain any characters, except double-quotes or newlines. Multiple attributes should be separated by space or tab characters.

Example:
```
some-tag -attr1="value"
person -firstName="John" -lastName="Doe"
```

#### Body
A body appears right after a tag identifier or after all attributes, followed by a space or tab character. A body begins with a left (opening) delimiter (default `<<`), followed by the content on the next line, and is closed by a right (closing) delimiter (default `>>`), which should always appear on the new line. The content can contain any text and characters.


Example:
```
tag << 
Some content
>>

tag -attr="value" <<
Something
>>
```

#### Inline body
As an alternative syntax, a body can as well stay on the same line as the tag, without moving content to the next line. Such a body is called an inline body and it can not contain any newline characters within the content. It follows the same syntax as a regular body, except the left and right delimiters are not separated by a newline character.

Example:
```
tag <<Inline body>>
```

### Custom delimiters
It is possible to change body delimiters via built-in tag `delimiters`. This tag must be defined at the top of the document before all other tags. It has two attributes `-left` for an opening delimiter, and `-right` for a closing delimiter. Both left and right delimiters can contain any sequence of UNICODE characters, except for invisible characters (e.g. whitespaces).

Example:
```
delimiters -left="{" -right="}"

tag {Something}
```

Also UNICODE characters are supported:
```
delimiters -left="👉" -right="👈"

tag 👉Some content👈
```