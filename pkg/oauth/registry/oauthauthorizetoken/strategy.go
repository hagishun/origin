package oauthauthorizetoken

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	kstorage "k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/util/validation/field"

	scopeauthorizer "github.com/openshift/origin/pkg/authorization/authorizer/scope"
	"github.com/openshift/origin/pkg/oauth/api"
	"github.com/openshift/origin/pkg/oauth/api/validation"
	"github.com/openshift/origin/pkg/oauth/registry/oauthclient"
)

// strategy implements behavior for OAuthAuthorizeTokens
type strategy struct {
	runtime.ObjectTyper

	clientGetter oauthclient.Getter
}

var _ rest.RESTCreateStrategy = strategy{}
var _ rest.RESTUpdateStrategy = strategy{}

func NewStrategy(clientGetter oauthclient.Getter) strategy {
	return strategy{ObjectTyper: kapi.Scheme, clientGetter: clientGetter}
}

func (strategy) PrepareForUpdate(ctx kapi.Context, obj, old runtime.Object) {}

// NamespaceScoped is false for OAuth objects
func (strategy) NamespaceScoped() bool {
	return false
}

func (strategy) GenerateName(base string) string {
	return base
}

func (strategy) PrepareForCreate(ctx kapi.Context, obj runtime.Object) {
}

// Canonicalize normalizes the object after validation.
func (strategy) Canonicalize(obj runtime.Object) {
}

// Validate validates a new token
func (s strategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	token := obj.(*api.OAuthAuthorizeToken)
	validationErrors := validation.ValidateAuthorizeToken(token)

	client, err := s.clientGetter.GetClient(ctx, token.ClientName)
	if err != nil {
		return append(validationErrors, field.InternalError(field.NewPath("clientName"), err))
	}
	if err := scopeauthorizer.ValidateScopeRestrictions(client, token.Scopes...); err != nil {
		return append(validationErrors, field.InternalError(field.NewPath("clientName"), err))
	}

	return validationErrors
}

// ValidateUpdate validates an update
func (s strategy) ValidateUpdate(ctx kapi.Context, obj, old runtime.Object) field.ErrorList {
	oldToken := old.(*api.OAuthAuthorizeToken)
	newToken := obj.(*api.OAuthAuthorizeToken)
	return validation.ValidateAuthorizeTokenUpdate(newToken, oldToken)
}

// AllowCreateOnUpdate is false for OAuth objects
func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) AllowUnconditionalUpdate() bool {
	return false
}

// GetAttrs returns labels and fields of a given object for filtering purposes
func GetAttrs(o runtime.Object) (labels.Set, fields.Set, error) {
	obj, ok := o.(*api.OAuthAuthorizeToken)
	if !ok {
		return nil, nil, fmt.Errorf("not a OAuthAuthorizeToken")
	}
	return labels.Set(obj.Labels), SelectableFields(obj), nil
}

// Matcher returns a generic matcher for a given label and field selector.
func Matcher(label labels.Selector, field fields.Selector) kstorage.SelectionPredicate {
	return kstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// SelectableFields returns a field set that can be used for filter selection
func SelectableFields(obj *api.OAuthAuthorizeToken) fields.Set {
	return api.OAuthAuthorizeTokenToSelectableFields(obj)
}
