package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-openapi/loads"
	"gopkg.in/yaml.v2"
)

//
// Generator
//

type Generator struct {
	ConfigurationPath string
	Configuration     Configuration
	OpenAPI           *loads.Document
	Nodes             Nodes
	Writer            *os.File
}

func NewGenerator(configurationPath string) (*Generator, error) {
	self := Generator{ConfigurationPath: configurationPath}
	if err := self.ReadConfig(); err == nil {
		if err = self.ReadOpenAPI(); err == nil {
			self.Nodes = self.ReadNodes()
			return &self, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Generator) OpenWriter(fileName string) error {
	var err error
	path := filepath.Join(filepath.Dir(self.ConfigurationPath), self.Configuration.OutputDir, fileName)
	if self.Writer, err = os.Create(path); err == nil {
		fmt.Println("writing", path)
	}
	return err
}

func (self *Generator) CloseWriter() error {
	if self.Writer != nil {
		return self.Writer.Close()
	} else {
		return nil
	}
}

func (self *Generator) Generate() error {
	capabilityNodes, dataNodes := self.SplitNodes()

	if err := self.GenerateFile("capabilities.yaml", "capability", capabilityNodes); err != nil {
		return err
	}

	if err := self.GenerateFile("data.yaml", "data", dataNodes); err != nil {
		return err
	}

	return nil
}

func (self *Generator) GenerateFile(fileName string, entity string, nodes Nodes) error {
	if err := self.OpenWriter(fileName); err != nil {
		return err
	}
	defer self.CloseWriter()

	self.WriteHeader(entity)

	self.Writeln()
	self.WriteKey(0, entity+"_types")

	isCapability := entity == "capability"
	for _, node := range nodes {
		self.Writeln()
		self.WriteType(node, isCapability)
	}

	return nil
}

func (self *Generator) ReadConfig() error {
	if bytes, err := os.ReadFile(self.ConfigurationPath); err == nil {
		return yaml.Unmarshal(bytes, &self.Configuration)
	} else {
		return err
	}
}

func (self *Generator) ReadOpenAPI() error {
	var err error
	self.OpenAPI, err = loads.JSONSpec(self.Configuration.OpenAPI)
	return err
}
