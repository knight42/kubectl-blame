package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Options struct {
	timeFormat string

	namespace string
	args      []string
	builder   *resource.Builder
}

var (
	blameLong = `Annotate each line in the given resource's YAML with information from the managedFields
to show who last modified the field.

As long as the field '.metadata.manageFields' of the given resource is set properly, this command
is able to display the manager of each field.`

	blameExample = `
# Blame pod 'foo' in default namespace
kubectl blame pods foo

# Blame deployment 'bar' in 'ns1' namespace
kubectl blame -n ns1 deploy bar

# Blame deployment 'bar' in 'ns1' namespace and hide the update time
kubectl blame -n ns1 --time none deploy bar
`
)

func CheckErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func NewCmdBlame() *cobra.Command {
	o := Options{}
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
	f.AddFlags(cmd.Flags())
	cmd.Flags().StringVar(&o.timeFormat, "time", TimeFormatRelative, "Time format. One of: full|relative|none.")
	return cmd
}

func (o *Options) Complete(f *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	var err error
	o.namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	o.args = args
	o.builder = resource.NewBuilder(f)
	return nil
}

func (o *Options) Validate() error {
	if len(o.args) != 2 {
		return fmt.Errorf("wrong number of arguments")
	}
	return nil
}

func (o *Options) Run() error {
	result := o.builder.Unstructured().
		NamespaceParam(o.namespace).DefaultNamespace().
		ResourceNames(o.args[0], o.args[1]).SingleResourceType().
		Latest().
		Do()
	if err := result.Err(); err != nil {
		return err
	}

	infos, err := result.Infos()
	if err != nil {
		return err
	}
	if len(infos) < 1 {
		return fmt.Errorf("not found")
	}

	info := infos[0]
	obj, ok := info.Object.(metav1.Object)
	if !ok {
		return fmt.Errorf("unsupported object: %v: %s/%s", info.Mapping.Resource, info.Namespace, info.Name)
	}

	data, err := MarshalMetaObject(obj, o.timeFormat)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}
