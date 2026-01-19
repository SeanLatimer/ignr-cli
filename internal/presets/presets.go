// Package presets provides preset management and suggestions.
package presets

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.seanlatimer.dev/ignr/internal/config"
	"gopkg.in/yaml.v3"
)

type Preset struct {
	Key       string   `yaml:"key,omitempty"`
	Name      string   `yaml:"name"`
	Templates []string `yaml:"templates"`
	Created   string   `yaml:"created"`
	Updated   string   `yaml:"updated"`
}

type PresetStore struct {
	Presets []Preset `yaml:"presets"`
}

func LoadPresets() (PresetStore, error) {
	path, err := config.GetPresetsPath()
	if err != nil {
		return PresetStore{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PresetStore{}, nil
		}
		return PresetStore{}, fmt.Errorf("read presets: %w", err)
	}

	var store PresetStore
	if err := yaml.Unmarshal(data, &store); err != nil {
		return PresetStore{}, fmt.Errorf("parse presets: %w", err)
	}
	for i := range store.Presets {
		if strings.TrimSpace(store.Presets[i].Key) == "" {
			store.Presets[i].Key = SluggifyName(store.Presets[i].Name)
		}
	}
	return store, nil
}

func SavePresets(store PresetStore) error {
	path, err := config.GetPresetsPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(store)
	if err != nil {
		return fmt.Errorf("marshal presets: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write presets: %w", err)
	}
	return nil
}

func FindPreset(name string) (Preset, bool, error) {
	store, err := LoadPresets()
	if err != nil {
		return Preset{}, false, err
	}
	targetKey := SluggifyName(name)
	for _, preset := range store.Presets {
		if strings.EqualFold(preset.Key, name) || strings.EqualFold(preset.Key, targetKey) {
			return preset, true, nil
		}
		if strings.EqualFold(preset.Name, name) {
			return preset, true, nil
		}
	}
	return Preset{}, false, nil
}

func CreatePreset(name string, templates []string) error {
	store, err := LoadPresets()
	if err != nil {
		return err
	}
	key := SluggifyName(name)
	for _, preset := range store.Presets {
		if strings.EqualFold(preset.Key, key) {
			return fmt.Errorf("preset key already exists: %s", key)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	store.Presets = append(store.Presets, Preset{
		Key:       key,
		Name:      name,
		Templates: templates,
		Created:   now,
		Updated:   now,
	})
	return SavePresets(store)
}

func EditPreset(name string, templates []string) error {
	store, err := LoadPresets()
	if err != nil {
		return err
	}

	index, ok := findPresetIndex(store, name)
	if !ok {
		return fmt.Errorf("preset not found: %s", name)
	}
	store.Presets[index].Templates = templates
	store.Presets[index].Updated = time.Now().UTC().Format(time.RFC3339)
	return SavePresets(store)
}

func DeletePreset(name string) error {
	store, err := LoadPresets()
	if err != nil {
		return err
	}

	index, ok := findPresetIndex(store, name)
	if !ok {
		return fmt.Errorf("preset not found: %s", name)
	}

	store.Presets = append(store.Presets[:index], store.Presets[index+1:]...)
	return SavePresets(store)
}

func ListPresets() ([]Preset, error) {
	store, err := LoadPresets()
	if err != nil {
		return nil, err
	}
	return store.Presets, nil
}

func findPresetIndex(store PresetStore, name string) (int, bool) {
	targetKey := SluggifyName(name)
	for i, preset := range store.Presets {
		if strings.EqualFold(preset.Key, name) || strings.EqualFold(preset.Key, targetKey) {
			return i, true
		}
		if strings.EqualFold(preset.Name, name) {
			return i, true
		}
	}
	return -1, false
}

func SluggifyName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	lastHyphen := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastHyphen = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case r == ' ' || r == '_':
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		case r == '-':
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		default:
			continue
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "preset"
	}
	return result
}
