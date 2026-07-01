package config

import (
	"fmt"
	"os"
	"strings"
)

// Validate checks that all paths referenced in the config exist on disk.
// All missing or mistyped paths are collected before returning so the caller
// sees the complete picture in one error message.
func (c *Config) Validate() error {
	var errs []string

	for _, item := range c.Content {
		if err := expectDir(item.Path, fmt.Sprintf("content %q", item.Name)); err != nil {
			errs = append(errs, err.Error())
		}
	}

	for _, item := range c.Data {
		if err := expectFile(item.Path, fmt.Sprintf("data %q path", item.Name)); err != nil {
			errs = append(errs, err.Error())
		}
		if item.Form != "" {
			if err := expectFile(item.Form, fmt.Sprintf("data %q form", item.Name)); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if item.Schema != "" {
			if err := expectFile(item.Schema, fmt.Sprintf("data %q schema", item.Name)); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if item.RecordTemplate != "" {
			if err := expectFile(item.RecordTemplate, fmt.Sprintf("data %q record_template", item.Name)); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

func expectDir(path, label string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: directory not found: %s", label, path)
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s: expected a directory, got a file: %s", label, path)
	}
	return nil
}

func expectFile(path, label string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: file not found: %s", label, path)
	}
	if fi.IsDir() {
		return fmt.Errorf("%s: expected a file, got a directory: %s", label, path)
	}
	return nil
}
