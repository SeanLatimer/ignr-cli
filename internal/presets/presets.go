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
	for _, preset := range store.Presets {
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
	for _, preset := range store.Presets {
		if strings.EqualFold(preset.Name, name) {
			return fmt.Errorf("preset already exists: %s", name)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	store.Presets = append(store.Presets, Preset{
		Name:      name,
		Templates: templates,
		Created:   now,
		Updated:   now,
	})
	return SavePresets(store)
}

func UpdatePreset(name string, templates []string) error {
	store, err := LoadPresets()
	if err != nil {
		return err
	}

	for i, preset := range store.Presets {
		if strings.EqualFold(preset.Name, name) {
			store.Presets[i].Templates = templates
			store.Presets[i].Updated = time.Now().UTC().Format(time.RFC3339)
			return SavePresets(store)
		}
	}
	return fmt.Errorf("preset not found: %s", name)
}

func DeletePreset(name string) error {
	store, err := LoadPresets()
	if err != nil {
		return err
	}

	filtered := make([]Preset, 0, len(store.Presets))
	found := false
	for _, preset := range store.Presets {
		if strings.EqualFold(preset.Name, name) {
			found = true
			continue
		}
		filtered = append(filtered, preset)
	}
	if !found {
		return fmt.Errorf("preset not found: %s", name)
	}

	store.Presets = filtered
	return SavePresets(store)
}

func ListPresets() ([]Preset, error) {
	store, err := LoadPresets()
	if err != nil {
		return nil, err
	}
	return store.Presets, nil
}
