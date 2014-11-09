package plugin

var plugins = make(map[string]Plugin)

func Add(name string, plugin Plugin) {
	if _, exists := plugins[name]; exists {
		panic("plugin already registered with name " + name)
	}
	plugins[name] = plugin
}

func All() map[string]Plugin {
	return plugins
}

type Plugin interface {
	Load(LoadContext) error
}

type LoadContext interface {
	Args() []string
	Register(name string, p Plugin) error
}
