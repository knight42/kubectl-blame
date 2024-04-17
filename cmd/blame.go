package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Options struct {
	timeFormat   string
	inputFile    string
	fileNameOpts resource.FilenameOptions

	namespace string
	enforceNs bool
	args      []string
	builder   *resource.Builder
	out       io.Writer
}

var (
	blameLong = `Annotate each line in the given resource's YAML with information from the managedFields
to show who last modified the field.

As long as the field '.metadata.manageFields' of the given resource is set properly, this command
is able to display the manager of each field.`

	blameExample = `
# Blame pod 'foo' in default namespace
kubectl blame pods foo

# Blame deployment 'foo' and 'bar' in 'ns1' namespace
kubectl blame -n ns1 deploy foo bar

# Blame deployment 'bar' in 'ns1' namespace and hide the update time
kubectl blame -n ns1 --time none deploy bar

# Blame resources in file 'pod.yaml'(will access remote server)
kubectl blame -f pod.yaml

# Blame deployment saved in local file 'deployment.yaml'(will NOT access remote server)
kubectl blame -i deployment.yaml
# Or
cat deployment.yaml | kubectl blame -i -
`
)

func CheckErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func NewCmdBlame() *cobra.Command {
	o := Options{out: os.Stdout}
	f := genericclioptions.NewConfigFlags(true)
	cmd := &cobra.Command{
		Use:                   "kubectl blame TYPE[.VERSION][.GROUP] NAME",
		DisableFlagsInUseLine: true,
		Short:                 "Show the manager of each field of given resource",
		Long:                  blameLong,
		Example:               blameExample,
		Run: func(cmd *cobra.Command, args []string) {
			CheckErr(o.Complete(f, cmd, args))
			CheckErr(o.Validate())
			CheckErr(o.Run())
		},
	}
	flags := cmd.Flags()
	f.AddFlags(flags)
	flags.StringVar(&o.timeFormat, "time", TimeFormatRelative, "Time format. One of: full|relative|none.")
	flags.StringSliceVarP(&o.fileNameOpts.Filenames, "filename", "f", o.fileNameOpts.Filenames, "Filename identifying the resource to get from a server.")
	flags.StringVarP(&o.inputFile, "input", "i", "", "Read object from the give file. When the file is -, read standard input.")
	return cmd
}

func (o *Options) Complete(f *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	var err error
	o.namespace, o.enforceNs, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	o.args = args
	o.builder = resource.NewBuilder(f)
	return nil
}

func (o *Options) Validate() error {
	return nil
}

func (o *Options) visitLocalObjects(visit func(object metav1.Object) error) error {
	// get objects from stdin or local files
	var input io.Reader
	if o.inputFile == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(o.inputFile)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
	}

	decoder := yaml.NewYAMLOrJSONDecoder(input, 4096)
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		err = visit(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) visitClusterObjects(visit func(object metav1.Object) error) error {
	result := o.builder.Unstructured().
		NamespaceParam(o.namespace).DefaultNamespace().
		FilenameParam(o.enforceNs, &o.fileNameOpts).
		ResourceTypeOrNameArgs(true, o.args...).
		Latest().
		Do()
	if err := result.Err(); err != nil {
		return err
	}

	return result.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		switch obj := info.Object.(type) {
		case metav1.Object:
			return visit(obj)
		case *unstructured.UnstructuredList:
			for _, item := range obj.Items {
				if err := visit(&item); err != nil {
					return err
				}
			}
			return nil
		default:
			return fmt.Errorf("unsupported object: %v: %s/%s", info.Mapping.Resource, info.Namespace, info.Name)
		}
	})
}

func (o *Options) Run() error {
	enc := newEncoder(o.out, func(obj metav1.Object) ([]byte, error) {
		return MarshalMetaObject(obj, o.timeFormat)
	})

	if len(o.inputFile) == 0 {
		return o.visitClusterObjects(enc.Encode)
	}

	return o.visitLocalObjects(enc.Encode)
}
