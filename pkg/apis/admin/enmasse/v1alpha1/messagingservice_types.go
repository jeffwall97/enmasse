package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AuthenticationSpec struct {
	Type            string `json:"type"`
	AdminSecretName string `json:"adminSecretName"`
}

type MonitoringSpec struct {
	Enabled bool `json:"enabled"`
}

// MessagingServiceSpec defines the desired state of MessagingService
type MessagingServiceSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Namespace      string               `json:"namespace,omitempty"`
	Authentication []AuthenticationSpec `json:"authentication,omitempty"`
	Monitoring     MonitoringSpec       `json:"monitoring,omitempty"`
}

// MessagingServiceStatus defines the observed state of MessagingService
type MessagingServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MessagingService is the Schema for the messagingservices API
// +k8s:openapi-gen=true
type MessagingService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MessagingServiceSpec   `json:"spec,omitempty"`
	Status MessagingServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MessagingServiceList contains a list of MessagingService
type MessagingServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MessagingService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MessagingService{}, &MessagingServiceList{})
}
