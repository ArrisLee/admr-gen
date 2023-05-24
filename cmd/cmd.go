package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"errors"

	"github.com/ghodss/yaml"
	admv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Params struct {
	YamlFile  string
	Operation string
	Output    string
}

func (params *Params) Validate() error {
	if params.YamlFile == "" {
		return errors.New("`file` parameter is mandatory, usage: --file=<path/to/yaml/file>")
	}
	if params.Operation == "" {
		params.Operation = "create"
	}
	if params.Operation != "" && (params.Operation != "create" && params.Operation != "update" && params.Operation != "delete") {
		return fmt.Errorf("invalid `operation` paramter: %s, usage: --operation=update or --operation=update or --operation=delete", params.Output)
	}
	if params.Output == "" {
		params.Output = "yaml"
	}
	if params.Output != "" && (params.Output != "json" && params.Output != "yaml") {
		return fmt.Errorf("invalid `output` paramter: %s, usage: --output=yaml or --output=json", params.Output)
	}
	return nil
}

func Run(params *Params) (string, error) {
	data, err := os.ReadFile(params.YamlFile)
	if err != nil {
		return "", err
	}

	objMap := make(map[string]interface{})
	err = yaml.Unmarshal(data, &objMap)
	if err != nil {
		return "", err
	}

	admissionReview, err := createAdmissionReview(objMap, data, params.Operation)
	if err != nil {
		return "", err
	}

	var output string

	if params.Output == "yaml" {
		raw, err := yaml.Marshal(admissionReview)
		if err != nil {
			return "", err
		}
		output = string(raw)
	} else {
		raw, err := json.Marshal(admissionReview)
		if err != nil {
			return "", err
		}
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, raw, "", "    "); err != nil {
			return "", err
		}
		output = prettyJSON.String()
	}

	return output, nil
}

func createAdmissionReview(objMap map[string]interface{}, data []byte, operation string) (*admv1.AdmissionReview, error) {
	op, err := getOperation(operation)
	if err != nil {
		return nil, err
	}

	apiVersion, ok := objMap["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve `apiVersion` from object")
	}

	kind, ok := objMap["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve `kind` from object")
	}

	gvk := metav1.GroupVersionKind{
		Group:   "",
		Version: apiVersion,
		Kind:    kind,
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %v", err)
	}

	dryRun := true

	result := &admv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/" + admv1.SchemeGroupVersion.Version,
		},
		Request: &admv1.AdmissionRequest{
			UID:         uuid.NewUUID(),
			Kind:        gvk,
			RequestKind: &gvk,
			Resource:    metav1.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: gvk.Kind},
			Operation:   op,
			UserInfo: v1.UserInfo{
				Username: "fake-k8s-admin-review",
				UID:      string(uuid.NewUUID()),
			},
			Object: runtime.RawExtension{
				Raw: jsonData,
			},
			DryRun: &dryRun,
		},
	}

	if op == admv1.Update || op == admv1.Delete {
		if op == admv1.Delete {
			objJSON, err := json.Marshal(objMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal object to JSON: %v", err)
			}
			result.Request.Object = runtime.RawExtension{}
			result.Request.OldObject = runtime.RawExtension{
				Raw: objJSON,
			}
		} else {
			oldObjMap := generateObjectForUpdateOperation(objMap)
			oldObjJSON, err := json.Marshal(oldObjMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal object to JSON: %v", err)
			}
			result.Request.OldObject = runtime.RawExtension{
				Raw: oldObjJSON,
			}
		}
	}

	return result, nil
}

func generateObjectForUpdateOperation(objMap map[string]interface{}) map[string]interface{} {
	oldObjMap := make(map[string]interface{}, len(objMap))
	for k, v := range objMap {
		oldObjMap[k] = v
	}

	metadata, ok := oldObjMap["metadata"].(map[string]interface{})
	if !ok {
		return oldObjMap
	}

	name, ok := metadata["name"].(string)
	if !ok {
		return oldObjMap
	}

	metadata["name"] = name + "-old"
	return oldObjMap
}

func getOperation(operation string) (admv1.Operation, error) {
	switch operation {
	case "create":
		return admv1.Create, nil
	case "update":
		return admv1.Update, nil
	case "delete":
		return admv1.Delete, nil
	default:
		return "", fmt.Errorf("invalid operation: %s", operation)
	}
}
