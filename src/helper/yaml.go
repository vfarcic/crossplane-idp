package helper

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		Compositions []Composition
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

	// tableRow struct {
	// 	API       string `header:"API"`
	// 	Name      string `header:"Name"`
	// 	ClaimName string `header:"Claim"`
	// }

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

	// CompositionSelector struct {
	// 	MatchLabels MatchLabels `yaml:"matchLabels"`
	// }

	// MatchLabels interface{}

	// WriteConnectionSecretToRef struct {
	// 	Name      string
	// 	Namespace string
	// }

	Properties interface{}
)

const insertHere = "INSERT_HERE"

var allCompositions = Compositions{}

var allXRDs = XRDs{}

func GetXRDs() XRDs {
	compositions := GetCompositions()
	if len(allXRDs.Items) == 0 {
		output, err := exec.Command("kubectl", "get", "compositeresourcedefinitions.apiextensions.crossplane.io", "-o", "yaml").Output()
		if err != nil {
			os.Stderr.WriteString(err.Error())
		}
		yaml.Unmarshal([]byte(string(output)), &allXRDs)
		for i, xrd := range allXRDs.Items {
			xrdCompositions := []Composition{}
			for _, composition := range compositions.Items {
				xrdGroupVersion := fmt.Sprintf("%s/%s", xrd.Spec.Group, xrd.Spec.Versions[0].Name)
				if xrdGroupVersion == composition.Spec.CompositeTypeRef.ApiVersion && xrd.Spec.Names.Kind == composition.Spec.CompositeTypeRef.Kind {
					xrdCompositions = append(xrdCompositions, composition)
				}
				allXRDs.Items[i].Compositions = xrdCompositions
			}
		}
	}
	return allXRDs
}

func GetXRD(name string) XRD {
	xrds := GetXRDs()
	for _, xrd := range xrds.Items {
		if xrd.Metadata.Name == name {
			return xrd
		}
	}
	return XRD{}
}

func GetCompositions() Compositions {
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

func GetXRYaml(xrdName, compositionName string, comments bool) string {
	crd := GetCRD(xrdName)
	xr := GetXR(crd, xrdName, compositionName, insertHere, comments)
	yamlData, err := yaml.Marshal(xr)
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	yaml := strings.ReplaceAll(string(yamlData), "'", "")
	return yaml
}

func GetXRYamlWithFields(xrdName, compositionName string) string {
	yaml := GetXRYaml(xrdName, compositionName, true)
	fieldsCount := strings.Count(yaml, insertHere)
	for i := 0; i < fieldsCount; i++ {
		field := fmt.Sprintf(
			`<input type="text" id="field%d" name="field%d">`,
			i,
			i,
		)
		yaml = strings.Replace(yaml, insertHere, field, 1)
	}
	yamlWithFields := fmt.Sprintf(
		`<form action="/xr"><pre>
%s
<input type="hidden" id="xrdName" name="xrdName" value="%s">
<input type="hidden" id="compositionName" name="compositionName" value="%s">
<input type="hidden" id="fieldsCount" name="fieldsCount" value="%d">
<input type="submit" value="Get YAML">
</pre></form>`,
		yaml,
		xrdName,
		compositionName,
		fieldsCount,
	)
	return yamlWithFields
}

func GetXRYamlWithValues(xrdName, compositionName string, fields []string) string {
	yaml := GetXRYaml(xrdName, compositionName, false)
	for _, field := range fields {
		yaml = strings.Replace(yaml, insertHere, field, 1)
	}
	yamlWithValues := fmt.Sprintf(
		`<pre>
%s
</pre>`,
		yaml,
	)
	return yamlWithValues
}

func GetCRD(api string) CRD {
	output, err := exec.Command("kubectl", "get", "crd", api, "-o", "yaml").Output()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	crd := CRD{}
	yaml.Unmarshal([]byte(string(output)), &crd)
	return crd
}

func GetXR(crd CRD, xrdName, compositionName, emptyValue string, addComments bool) XR {
	xr := XR{}
	xr.ApiVersion = crd.Spec.Group + "/" + crd.Spec.Versions[0].Name
	xr.Kind = crd.Spec.Names.Kind
	xr.Metadata.Name = emptyValue
	spec := make(map[interface{}]interface{})
	properties := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties.Spec.Properties.(map[interface{}]interface{})
	for key, value := range properties {
		switch key {
		case "claimRef", "compositionUpdatePolicy", "resourceRefs", "compositionRef", "compositionRevisionRef", "publishConnectionDetailsTo":
			// Ignore
		case "compositionSelector":
			labels := make(map[string]string)
			if len(compositionName) == 0 {
				labels["SOME_KEY"] = emptyValue
				labels["SOME_OTHER_KEY"] = emptyValue
			} else {
				xrd := GetXRD(xrdName)
				for _, composition := range xrd.Compositions {
					if composition.Metadata.Name == compositionName {
						for key, value := range composition.Metadata.Labels {
							labels[key] = value
						}
					}
				}
			}
			matchLabels := make(map[string]interface{})
			matchLabels["matchLabels"] = labels
			spec["CompositionSelector"] = matchLabels
		case "writeConnectionSecretToRef":
			secrets := make(map[string]string)
			secrets["name"] = emptyValue
			secrets["namespace"] = emptyValue
			if addComments {
				secrets["name"] = secrets["name"] + " # The name of the secret with authentication (string)"
				secrets["namespace"] = secrets["namespace"] + " # The namespace for the secret (string)"
			}
			spec["writeConnectionSecretToRef"] = secrets
		default:
			switch v := value.(type) {
			case map[interface{}]interface{}:
				subKey := fmt.Sprintf("%v", key)
				spec[subKey] = processMapInterface(v, addComments)
			}
		}
		xr.Spec = spec
	}
	return xr
}

func processMapInterface(properties interface{}, addComments bool) interface{} {
	var subProperties interface{}
	hasProperties := false
	description := ""
	xDefault := ""
	xType := ""
	for key, value := range properties.(map[interface{}]interface{}) {
		if key == "properties" {
			hasProperties = true
			subProperties = value
		}
		keyString := fmt.Sprintf("%v", key)
		valueString := fmt.Sprintf("%v", value)
		switch keyString {
		case "description":
			description = valueString
		case "type":
			xType = valueString
		case "default":
			xDefault = valueString
		}
	}
	if hasProperties {
		newSubProperties := make(map[interface{}]interface{})
		for key, value := range subProperties.(map[interface{}]interface{}) {
			newSubProperties[key] = processMapInterface(value, addComments)
		}
		return newSubProperties
	}
	comment := ""
	if addComments {
		info := fmt.Sprintf("Type: %s", xType)
		if len(xDefault) > 0 {
			info = fmt.Sprintf("%s; Default: %s", info, xDefault)
		}
		comment = fmt.Sprintf(
			" # %s (%s)",
			description,
			info,
		)
	}
	return fmt.Sprintf(
		"%s%s",
		insertHere,
		comment,
	)
}
