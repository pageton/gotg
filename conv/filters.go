package conv

import "github.com/pageton/gotg/types"

var Filters = struct {
	Text  Filter
	Photo Filter
	Video Filter
	Audio Filter
	Voice Filter
	Media Filter
	Any   Filter
}{
	Text: func(m *types.Message) bool {
		return m != nil && m.Message != nil && m.Text != "" && m.Media == nil
	},
	Photo: func(m *types.Message) bool {
		if m == nil || m.Message == nil || m.Media == nil {
			return false
		}
		ph := m.Photo()
		return ph != nil
	},
	Video: func(m *types.Message) bool {
		return m != nil && m.Message != nil && m.Video() != nil
	},
	Audio: func(m *types.Message) bool {
		return m != nil && m.Message != nil && m.Audio() != nil
	},
	Voice: func(m *types.Message) bool {
		return m != nil && m.Message != nil && m.Voice() != nil
	},
	Media: func(m *types.Message) bool {
		return m != nil && m.Message != nil && m.Media != nil
	},
	Any: func(m *types.Message) bool {
		return m != nil
	},
}
