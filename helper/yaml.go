package helper

import (
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type (
	Compositions struct {
		Items []Composition
	}

	Composition struct {
		Metadata struct {
			Name   string
			Labels map[string]string
		}
		Spec struct {
			CompositeTypeRef CompositeTypeRef `yaml:"compositeTypeRef"`
		}
	}

	XRDs struct {
		Items []XRD
	}

	XRD struct {
		Metadata struct {
			Name string
		}
		Spec struct {
			Group      string
			ClaimNames KindPlural `yaml:"claimNames"`
			Names      KindPlural
			Versions   []Version `yaml:"versions"`
		}
	}

	XR struct {
		ApiVersion string `yaml:"apiVersion"`
		Kind       string
		Metadata   struct {
			Name string
		}
		Spec interface{}
	}

	Version struct {
		Name string
	}

	CompositeTypeRef struct {
		ApiVersion string `yaml:"apiVersion"`
		Kind       string
	}

	tableRow struct {
		API       string `header:"API"`
		Name      string `header:"Name"`
		ClaimName string `header:"Claim"`
	}

	KindPlural struct {
		Kind   string
		Plural string
	}

	CRD struct {
		ApiVersion string  `yaml:"apiVersion"`
		Spec       CrdSpec `yaml:"spec"`
	}

	CrdSpec struct {
		Group    string
		Versions []struct {
			Name   string
			Schema struct {
				OpenAPIV3Schema OpenAPIV3Schema `yaml:"openAPIV3Schema"`
			}
		}
		Names struct {
			Kind string
		}
	}

	OpenAPIV3Schema struct {
		Properties struct {
			Spec struct {
				Properties Properties
			}
		}
	}

	CompositionSelector struct {
		MatchLabels MatchLabels `yaml:"matchLabels"`
	}

	MatchLabels interface{}

	WriteConnectionSecretToRef struct {
		Name      string
		Namespace string
	}

	Properties interface{}
)

var allCompositions = Compositions{}

func getAllCompositions() Compositions {
	if len(allCompositions.Items) == 0 {
		allCompositions = Compositions{}
		yamlOutput, err := exec.Command("kubectl", "get", "compositions.apiextensions.crossplane.io", "-o", "yaml").Output()
		if err != nil {
			os.Stderr.WriteString(err.Error())
		}
		yaml.Unmarshal([]byte(string(yamlOutput)), &allCompositions)
	}
	return allCompositions
}
