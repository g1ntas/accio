package generator

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

func (g *Generator) readConfig(b []byte) error {
	err := toml.Unmarshal(b, &g)
	if err != nil {
		return err
	}
	if err := g.validate(); err != nil {
		return err
	}
	return nil
}

func (m PromptMap) UnmarshalTOML(data interface{}) error {
	prompts := data.(map[string]interface{})
	for key, def := range prompts {
		mapping := def.(map[string]interface{})
		typ, ok := mapping["type"].(string)
		if !ok {
			return fmt.Errorf("prompt %q has no type", key)
		}
		base := base{}
		base.msg, _ = mapping["message"].(string)
		base.help, _ = mapping["help"].(string)
		switch typ {
		case promptInput:
			m[key] = &input{base}
		case promptInteger:
			m[key] = &integer{base}
		case promptConfirm:
			m[key] = &confirm{base}
		case promptList:
			m[key] = &list{base}
		case promptChoice:
			m[key] = &choice{base, parseOptions(mapping)}
		case promptMultiChoice:
			m[key] = &multiChoice{base, parseOptions(mapping)}
		default:
			return fmt.Errorf("prompt %q with unknown type %q", key, typ)
		}
	}
	return nil
}

// parseOptions returns Prompt options parsed from given toml mapping
func parseOptions(tree map[string]interface{}) (options []string) {
	if opts, ok := tree["options"].([]interface{}); ok {
		options = make([]string, len(opts))
		for i, v := range opts {
			options[i] = v.(string)
		}
	}
	return
}

func (g *Generator) validate() error {
	return nil
}