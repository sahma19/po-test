package tests

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func RunUnitTests(testFiles []string) error {
	var originalRules []*filenameAndData

	for _, testFile := range testFiles {
		unitTestInp, readErr := parseTestFile(testFile)
		if readErr != nil {
			return readErr
		}

		for _, rulesFile := range unitTestInp.RuleFiles {
			processErr := processRulesFile(testFile, rulesFile, &originalRules)
			if processErr != nil {
				return processErr
			}
		}
	}

	promtoolErr := runPromtoolTests(testFiles)
	if promtoolErr != nil {
		restoreOriginalFiles(originalRules)
		return promtoolErr
	}

	restoreOriginalFiles(originalRules)
	return nil
}

func parseTestFile(testFile string) (*unitTestFile, error) {
	data, readErr := os.ReadFile(testFile)
	if readErr != nil {
		return nil, readErr
	}

	var unitTestInp unitTestFile
	unmarshalErr := yaml.Unmarshal(data, &unitTestInp)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &unitTestInp, nil
}

func processRulesFile(testFile, rulesFile string, originalRules *[]*filenameAndData) error {
	relativePath := fmt.Sprintf("%s/%s", filepath.Dir(testFile), rulesFile)

	yamlData, readErr := os.ReadFile(relativePath)
	if readErr != nil {
		return readErr
	}

	unstructured, unmarshalErr := unmarshalYAML(yamlData)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	if spec, found := unstructured["spec"]; found {
		updateErr := updateRulesFile(relativePath, yamlData, spec, originalRules)
		if updateErr != nil {
			return updateErr
		}
	} else {
		log.Printf("No spec found in file %s", rulesFile)
	}
	return nil
}

func unmarshalYAML(data []byte) (map[interface{}]interface{}, error) {
	unstructured := make(map[interface{}]interface{})
	unmarshalErr := yaml.Unmarshal(data, &unstructured)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return unstructured, nil
}

func updateRulesFile(filePath string, originalData []byte, spec interface{}, originalRules *[]*filenameAndData) error {
	ruleFileContentWithoutMetadata, marshalErr := yaml.Marshal(spec)
	if marshalErr != nil {
		return marshalErr
	}

	*originalRules = append(*originalRules, &filenameAndData{filePath, originalData})

	writeErr := os.WriteFile(filePath, ruleFileContentWithoutMetadata, 0o600)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func runPromtoolTests(testFiles []string) error {
	promtoolArgs := append([]string{"test", "rules"}, testFiles...)
	command := exec.Command("promtool", promtoolArgs...)
	output, execErr := command.CombinedOutput()
	log.Printf("%s", output)
	return execErr
}

func restoreOriginalFiles(rules []*filenameAndData) {
	for _, nameAndData := range rules {
		writeErr := os.WriteFile(nameAndData.filename, nameAndData.data, 0o600)
		if writeErr != nil {
			log.Fatalf("Failed to write file: %v", writeErr)
		}
	}
}

type filenameAndData struct {
	filename string
	data     []byte
}

type unitTestFile struct {
	RuleFiles []string `yaml:"rule_files"`
}
