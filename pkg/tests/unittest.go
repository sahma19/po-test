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
		unitTestInp, err := parseTestFile(testFile)
		if err != nil {
			return err
		}

		for _, rulesFile := range unitTestInp.RuleFiles {
			if err := processRulesFile(testFile, rulesFile, &originalRules); err != nil {
				return err
			}
		}
	}

	if err := runPromtoolTests(testFiles); err != nil {
		restoreOriginalFiles(originalRules)
		return err
	}

	restoreOriginalFiles(originalRules)
	return nil
}

func parseTestFile(testFile string) (*unitTestFile, error) {
	data, err := os.ReadFile(testFile)
	if err != nil {
		return nil, err
	}

	var unitTestInp unitTestFile
	if err := yaml.Unmarshal(data, &unitTestInp); err != nil {
		return nil, err
	}

	return &unitTestInp, nil
}

func processRulesFile(testFile, rulesFile string, originalRules *[]*filenameAndData) error {
	relativePath := fmt.Sprintf("%s/%s", filepath.Dir(testFile), rulesFile)

	yamlData, err := os.ReadFile(relativePath)
	if err != nil {
		return err
	}

	unstructured, err := unmarshalYAML(yamlData)
	if err != nil {
		return err
	}

	if spec, found := unstructured["spec"]; found {
		if err := updateRulesFile(relativePath, yamlData, spec, originalRules); err != nil {
			return err
		}
	} else {
		log.Printf("No spec found in file %s", rulesFile)
	}

	return nil
}

func unmarshalYAML(data []byte) (map[interface{}]interface{}, error) {
	unstructured := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, &unstructured); err != nil {
		return nil, err
	}
	return unstructured, nil
}

func updateRulesFile(filePath string, originalData []byte, spec interface{}, originalRules *[]*filenameAndData) error {
	ruleFileContentWithoutMetadata, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	*originalRules = append(*originalRules, &filenameAndData{filePath, originalData})

	if err := os.WriteFile(filePath, ruleFileContentWithoutMetadata, 0o600); err != nil {
		return err
	}

	return nil
}

func runPromtoolTests(testFiles []string) error {
	promtoolArgs := append([]string{"test", "rules"}, testFiles...)
	command := exec.Command("promtool", promtoolArgs...)
	output, err := command.CombinedOutput()
	log.Printf("%s", output)
	return err
}

func restoreOriginalFiles(rules []*filenameAndData) {
	for _, nameAndData := range rules {
		if err := os.WriteFile(nameAndData.filename, nameAndData.data, 0o600); err != nil {
			log.Fatalf("Failed to write file: %v", err)
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
