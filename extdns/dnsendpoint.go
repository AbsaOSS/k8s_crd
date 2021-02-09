package dnsendpoint

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/external-dns/endpoint"
)

const GroupName = "externaldns.k8s.io"
const GroupVersion = "v1alpha1"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

type ExtDNSClient struct {
	restClient rest.Interface
}

type ExtDNSInterface interface {
	DNSEndpoints(namespace string) DNSEndpoint
}

type dnsendpointClient struct {
	restClient rest.Interface
	ns         string
}

func (c *ExtDNSClient) DNSEndpoints(namespace string) DNSEndpoint {
	return &dnsendpointClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}

type DNSEndpoint interface {
	List(ctx context.Context, opts metav1.ListOptions) (*endpoint.DNSEndpointList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

func NewForConfig(c *rest.Config) (*ExtDNSClient, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &ExtDNSClient{restClient: client}, nil
}

func (c *dnsendpointClient) List(ctx context.Context, opts metav1.ListOptions) (*endpoint.DNSEndpointList, error) {
	result := endpoint.DNSEndpointList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("dnsendpoints").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *dnsendpointClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("dnsendpoints").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&endpoint.DNSEndpoint{},
		&endpoint.DNSEndpointList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}