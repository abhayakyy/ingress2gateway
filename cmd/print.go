/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// outputFormat contains currently set output format. Value assigned via --output/-o flag.
	// Defaults to YAML.
	outputFormat = "yaml"

	// namespace contains currently set namespace that should be used. Value assigned via
	// --namespace/-n flag.
	// Default behavior is to use the current namespace the user is in.
	namespace string

	// allNamespaces indicates whether all namespaces should be used. Value assigned via
	// --all-namespaces/-A flag.
	// If present, overrides the namespace variable.
	allNamespaces bool
)

// printCmd represents the print command. It prints HTTPRoutes and Gateways
// generated from Ingress resources.
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Prints HTTPRoutes and Gateways generated from Ingress resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		resourcePrinter, err := getResourcePrinter(outputFormat)
		if err != nil {
			return err
		}
		namespaceFilter, err := getNamespaceFilter(namespace, allNamespaces)
		if err != nil {
			return err
		}
		i2gw.Run(resourcePrinter, namespaceFilter)
		return nil
	},
}

// getResourcePrinter returns a specific type of printers.ResourcePrinter
// based on the provided outputFormat.
func getResourcePrinter(outputFormat string) (printers.ResourcePrinter, error) {
	switch outputFormat {
	case "yaml", "":
		return &printers.YAMLPrinter{}, nil
	case "json":
		return &printers.JSONPrinter{}, nil
	default:
		return nil, fmt.Errorf("%s is not a supported output format", outputFormat)
	}
}

// getNamespaceFilter returns a namespace filter, taking into consideration whether a specific
// namespace is requested, or all of them are.
func getNamespaceFilter(requestedNamespace string, useAllNamespaces bool) (string, error) {

	// When we should use all namespaces, return an empty string.
	// This is the first condition since it should override the requestedNamespace,
	// if specified.
	if useAllNamespaces {
		return "", nil
	}

	if requestedNamespace == "" {
		return getNamespaceInCurrentContext()
	}
	return requestedNamespace, nil
}

// getNamespaceInCurrentContext returns the namespace in the current active context of the user.
func getNamespaceInCurrentContext() (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	currentNamespace, _, err := kubeConfig.Namespace()

	return currentNamespace, err
}

func init() {
	var printFlags genericclioptions.JSONYamlPrintFlags
	allowedFormats := printFlags.AllowedFormats()

	printCmd.Flags().StringVarP(&outputFormat, "output", "o", "yaml",
		fmt.Sprintf(`Output format. One of: (%s)`, strings.Join(allowedFormats, ", ")))

	printCmd.Flags().StringVarP(&namespace, "namespace", "n", "",
		fmt.Sprintf(`If present, the namespace scope for this CLI request`))

	printCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false,
		fmt.Sprintf(`If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even
if specified with --namespace.`))

	rootCmd.AddCommand(printCmd)
}
