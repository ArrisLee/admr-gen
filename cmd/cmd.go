package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"errors"

	"github.com/ghodss/yaml"
	admv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Constants for operations and output formats
const (
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"

	OutputYAML = "yaml"
	OutputJSON = "json"

	DefaultOperation = OperationCreate
	DefaultOutput    = OutputYAML

	FakeUsername    = "fake-k8s-admin-review"
	OldObjectSuffix = "-old"
)

// Params holds the configuration for admission review generation
type Params struct {
	YamlFile  string
	Operation string
	Output    string
}

// Validate validates the parameters and sets defaults
func (params *Params) Validate() error {
	if params.YamlFile == "" {
		return errors.New("`file` parameter is mandatory, usage: --file=<path/to/yaml/file>")
	}

	if params.Operation == "" {
		params.Operation = DefaultOperation
	}
	if !isValidOperation(params.Operation) {
		return fmt.Errorf("invalid `operation` parameter: %s, usage: --operation=create or --operation=update or --operation=delete", params.Operation)
	}

	if params.Output == "" {
		params.Output = DefaultOutput
	}
	if !isValidOutput(params.Output) {
		return fmt.Errorf("invalid `output` parameter: %s, usage: --output=yaml or --output=json", params.Output)
	}

	return nil
}

// isValidOperation checks if the operation is valid
func isValidOperation(operation string) bool {
	switch operation {
	case OperationCreate, OperationUpdate, OperationDelete:
		return true
	default:
		return false
	}
}

// isValidOutput checks if the output format is valid
func isValidOutput(output string) bool {
	switch output {
	case OutputYAML, OutputJSON:
		return true
	default:
		return false
	}
}

// Run generates an admission review from the given parameters
func Run(params *Params) (string, error) {
	// Read and parse the YAML file
	kubeObject, err := readAndParseYAML(params.YamlFile)
	if err != nil {
		return "", fmt.Errorf("failed to read YAML file: %v", err)
	}

	// Create admission review
	admissionReview, err := createAdmissionReview(kubeObject, params.Operation)
	if err != nil {
		return "", fmt.Errorf("failed to create admission review: %v", err)
	}

	// Format output
	output, err := formatOutput(admissionReview, params.Output)
	if err != nil {
		return "", fmt.Errorf("failed to format output: %v", err)
	}

	return output, nil
}

// KubeObject represents a parsed Kubernetes object
type KubeObject struct {
	APIVersion string
	Kind       string
	ObjectMap  map[string]interface{}
	RawData    []byte
}

// readAndParseYAML reads and parses a YAML file into a KubeObject
func readAndParseYAML(filename string) (*KubeObject, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	objMap := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &objMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	apiVersion, ok := objMap["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve `apiVersion` from object or it's not a string")
	}

	kind, ok := objMap["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve `kind` from object or it's not a string")
	}

	return &KubeObject{
		APIVersion: apiVersion,
		Kind:       kind,
		ObjectMap:  objMap,
		RawData:    data,
	}, nil
}

// createAdmissionReview creates an admission review from a Kubernetes object
func createAdmissionReview(kubeObj *KubeObject, operation string) (*admv1.AdmissionReview, error) {
	op, err := getOperation(operation)
	if err != nil {
		return nil, err
	}

	gvk, err := parseGroupVersionKind(kubeObj.APIVersion, kubeObj.Kind)
	if err != nil {
		return nil, err
	}

	jsonData, err := yaml.YAMLToJSON(kubeObj.RawData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %v", err)
	}

	dryRun := true
	request := &admv1.AdmissionRequest{
		UID:         uuid.NewUUID(),
		Kind:        gvk,
		RequestKind: &gvk,
		Resource:    getResourceFromKind(gvk),
		Operation:   op,
		UserInfo: v1.UserInfo{
			Username: FakeUsername,
			UID:      string(uuid.NewUUID()),
		},
		Object: runtime.RawExtension{
			Raw: jsonData,
		},
		DryRun: &dryRun,
	}

	// Handle operations that need OldObject
	if err := setOldObject(request, kubeObj, op); err != nil {
		return nil, err
	}

	return &admv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/" + admv1.SchemeGroupVersion.Version,
		},
		Request: request,
	}, nil
}

// parseGroupVersionKind parses apiVersion and kind into GroupVersionKind
func parseGroupVersionKind(apiVersion, kind string) (metav1.GroupVersionKind, error) {
	var group, version string

	if strings.Contains(apiVersion, "/") {
		parts := strings.SplitN(apiVersion, "/", 2)
		group = parts[0]
		version = parts[1]
	} else {
		// Core API group (e.g., v1)
		group = ""
		version = apiVersion
	}

	return metav1.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}, nil
}

// getResourceFromKind converts a Kind to its corresponding Resource
func getResourceFromKind(gvk metav1.GroupVersionKind) metav1.GroupVersionResource {
	// Basic pluralization logic - in a real implementation, you'd want
	// a more sophisticated mapping or use the discovery API
	resource := strings.ToLower(gvk.Kind)

	// Handle common irregular plurals
	switch resource {
	case "policy":
		resource = "policies"
	case "networkpolicy":
		resource = "networkpolicies"
	case "ingress":
		resource = "ingresses"
	default:
		// Simple pluralization
		if strings.HasSuffix(resource, "s") {
			resource = resource + "es"
		} else if strings.HasSuffix(resource, "y") {
			resource = resource[:len(resource)-1] + "ies"
		} else {
			resource = resource + "s"
		}
	}

	return metav1.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resource,
	}
}

// setOldObject sets the OldObject field for update and delete operations
func setOldObject(request *admv1.AdmissionRequest, kubeObj *KubeObject, op admv1.Operation) error {
	if op != admv1.Update && op != admv1.Delete {
		return nil
	}

	var oldObjectMap map[string]interface{}

	if op == admv1.Delete {
		// For delete operations, the object being deleted is the oldObject
		oldObjectMap = kubeObj.ObjectMap
		request.Object = runtime.RawExtension{} // Clear the object for delete
	} else {
		// For update operations, create a modified version for oldObject
		oldObjectMap = generateOldObjectForUpdate(kubeObj.ObjectMap)
	}

	oldObjJSON, err := json.Marshal(oldObjectMap)
	if err != nil {
		return fmt.Errorf("failed to marshal old object to JSON: %v", err)
	}

	request.OldObject = runtime.RawExtension{
		Raw: oldObjJSON,
	}

	return nil
}

// generateOldObjectForUpdate creates a modified version of the object for update operations
func generateOldObjectForUpdate(objMap map[string]interface{}) map[string]interface{} {
	// Create a deep copy
	oldObjMap := make(map[string]interface{})
	for k, v := range objMap {
		oldObjMap[k] = v
	}

	// Modify the name to simulate an old version
	if metadata, ok := oldObjMap["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			metadata["name"] = name + OldObjectSuffix
		}
	}

	return oldObjMap
}

// formatOutput formats the admission review in the specified format
func formatOutput(admissionReview *admv1.AdmissionReview, format string) (string, error) {
	switch format {
	case OutputYAML:
		raw, err := yaml.Marshal(admissionReview)
		if err != nil {
			return "", fmt.Errorf("failed to marshal to YAML: %v", err)
		}
		return string(raw), nil

	case OutputJSON:
		raw, err := json.Marshal(admissionReview)
		if err != nil {
			return "", fmt.Errorf("failed to marshal to JSON: %v", err)
		}

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, raw, "", "    "); err != nil {
			return "", fmt.Errorf("failed to format JSON: %v", err)
		}
		return prettyJSON.String(), nil

	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// getOperation converts a string operation to admv1.Operation
func getOperation(operation string) (admv1.Operation, error) {
	switch operation {
	case OperationCreate:
		return admv1.Create, nil
	case OperationUpdate:
		return admv1.Update, nil
	case OperationDelete:
		return admv1.Delete, nil
	default:
		return "", fmt.Errorf("invalid operation: %s", operation)
	}
}
