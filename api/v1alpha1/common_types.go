package v1alpha1

type MariaDBStatusType string

const (
	MariaDBStatusReady MariaDBStatusType = "Ready"
	MariaDBStatusError MariaDBStatusType = "Error"
)
