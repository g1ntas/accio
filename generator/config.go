package generator

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"unicode"
)

func (g *Generator) readConfig(b []byte) error {
	err := toml.Unmarshal(b, &g)
	if err != nil {
		return err
	}
	return nil
}

func (m PromptMap) UnmarshalTOML(data interface{}) error {
	prompts := data.(map[string]interface{})
	for key, def := range prompts {
		if err := validatePromptKey(key); err != nil {
			return err
		}
		mapping := def.(map[string]interface{})
		typ, ok := mapping["type"].(string)
		if !ok {
			return fmt.Errorf("missing type for prompt %q", key)
		}
		base := Base{}
		var err error
		if base.Msg, err = parsePromptMessage(mapping, key); err != nil {
			return err
		}
		base.HelpText, _ = mapping["help"].(string)
		switch typ {
		case promptInput:
			m[key] = &input{base}
		case promptInteger:
			m[key] = &integer{base}
		case promptConfirm:
			m[key] = &confirm{base}
		case promptList:
			m[key] = &list{base}
		case promptChoice, promptMultiChoice:
			opts, err := parsePromptOptions(mapping, key)
			if err != nil {
				return err
			}
			if typ == promptChoice {
				m[key] = &choice{base, opts}
			} else {
				m[key] = &multiChoice{base, opts}
			}
		default:
			return fmt.Errorf("unknown type %q in prompt %q", typ, key)
		}
	}
	return nil
}

func parsePromptOptions(conf map[string]interface{}, key string) (options []string, err error) {
	opts, ok := conf["options"].([]interface{})
	if !ok || len(opts) == 0 {
		return []string{}, fmt.Errorf("no options were specified for prompt %q", key)
	}
	options = make([]string, len(opts))
	for i, v := range opts {
		options[i] = v.(string)
	}
	return
}

func validatePromptKey(k string) error {
	if len(k) > 64 {
		return fmt.Errorf("prompt key %q is too long, it must be not longer than 64 characters", k)
	}
	if isDigit(rune(k[0])) {
		return fmt.Errorf("prompt key %q should start with a letter or underscore, but got digit instead", k)
	}
	for _, r := range k {
		if !isLetter(r) && !isDigit(r) && r != '_' {
			return fmt.Errorf("prompt key %q contains invalid character %q", k, r)
		}
	}
	return nil
}

func parsePromptMessage(conf map[string]interface{}, key string) (string, error) {
	s, _ := conf["message"].(string)
	if len(s) == 0 {
		return "", fmt.Errorf("required setting 'message' is missing for prompt %q", key)
	}
	if len(s) > 128 {
		return "", fmt.Errorf("prompt message for %q is too long, it must be not longer than 128 characters", key)
	}
	return s, nil
}

// startsOrEndsWithRune checks whether string starts or ends with a given rune.
func startsOrEndsWithRune(s string, r rune) bool {
	return s[0:1] == string(r) || s[len(s)-1:] == string(r)
}

// isLetter checks whether r is an ASCII valid letter ([a-zA-Z]).
func isLetter(r rune) bool {
	return r <= unicode.MaxASCII && unicode.IsLetter(r)
}

// isDigit checks whether r is an ASCII valid numeric digit ([0-9]).
func isDigit(r rune) bool {
	return r <= unicode.MaxASCII && unicode.IsDigit(r)
}
